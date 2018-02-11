package main

import (
	"github.com/shirou/gopsutil/process"
	"os"
	"os/exec"
	"log"
	"time"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/faiface/beep/effects"
)

const maxVolume = 8 // 2^3

func fatalIfFailed(err error) {
	if err != nil {
		panic(err)
	}
}

func calculateCompleteUtilization(rootProcess *process.Process) float64 {
	cpuPercent, err := rootProcess.CPUPercent()
	if err != nil {
		return 0
	}

	completeUtilization := cpuPercent

	children, _ := rootProcess.Children()

	if children != nil {
		for _, childProcess := range children {
			completeUtilization += calculateCompleteUtilization(childProcess)
		}
	}

	return completeUtilization
}

func watchProcess(pid int, done chan interface{}, effectBuffer *beep.Buffer) {
	watchedProcess, err := process.NewProcess(int32(pid))
	fatalIfFailed(err)

	streamer := InfiniteLoop(effectBuffer.Streamer(0, effectBuffer.Len()))
	varyingSoundEffect := effects.Volume{ Streamer: streamer, Base: 1, Volume: 1, Silent: false }

	speaker.Init(effectBuffer.Format().SampleRate, effectBuffer.Format().SampleRate.N(time.Second/10))
	speaker.Play(&varyingSoundEffect)

	maxUtilization := 1.0

	shouldRun := true
	for shouldRun {
		utilization := calculateCompleteUtilization(watchedProcess)

		select {
			case <- done:
				shouldRun = false
			default:
				time.Sleep(1000 * time.Millisecond)
		}

		if maxUtilization == 0.0 || maxUtilization < utilization {
			maxUtilization = utilization
		}

		newVolume := 1.0 + ((utilization / maxUtilization) * maxVolume)

		speaker.Lock()
		varyingSoundEffect.Base = newVolume
		speaker.Unlock()
	}
}

func main() {
	if len(os.Args) > 1 {
		programName := os.Args[1]
		programParams := os.Args[2:]

		cmd := exec.Command(programName, programParams...)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		fatalIfFailed(err)

		soundEffect, format, err := wav.Decode(openAsset("data/sizzle.wav"))
		fatalIfFailed(err)

		effectBuffer := beep.NewBuffer(format)
		effectBuffer.Append(soundEffect)

		soundEffect.Close()

		done := make(chan interface{}, 1)

		go watchProcess(cmd.Process.Pid, done, effectBuffer)

		cmd.Wait()

		done <- nil
	} else {
		log.Fatal("Please specify a program")
	}
}
