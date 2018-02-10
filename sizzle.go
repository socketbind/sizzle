package main

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"os"
	"os/exec"
	"log"
	"time"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"io"
	"github.com/faiface/beep/speaker"
)

type ReadableClosableBytes struct {
	s        []byte
	i        int64 // current reading index
	prevRune int   // index of previous rune; or < 0
}

func (r *ReadableClosableBytes) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

func (r *ReadableClosableBytes) Close() error {
	return nil
}

func openAsset(name string) *ReadableClosableBytes {
	asset, err := Asset(name)
	fatalIfFailed(err)
	return &ReadableClosableBytes{ s: asset, i: 0, prevRune: -1 }
}

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

	/*processName, err := rootProcess.Name()
	if err != nil {
		return 0
	}

	fmt.Printf("%s CPU: %f\n", processName, cpuPercent)*/

	completeUtilization := cpuPercent



	children, _ := rootProcess.Children()

	if children != nil {
		for _, childProcess := range children {
			completeUtilization += calculateCompleteUtilization(childProcess)
		}
	}

	return completeUtilization
}

func watchProcess(pid int, done chan bool, effect *beep.StreamSeekCloser, format *beep.Format) {
	watchedProcess, err := process.NewProcess(int32(pid))
	fatalIfFailed(err)

	shouldRun := true

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(*effect)

	for shouldRun {
		utilization := calculateCompleteUtilization(watchedProcess)
		fmt.Printf("Complete utilization: %f\n", utilization)

		select {
			case <- done:
				shouldRun = false
			default:
				time.Sleep(1000 * time.Millisecond)
		}
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

		soundEffect, format, err := mp3.Decode(openAsset("data/sizzle.mp3"))
		fatalIfFailed(err)

		defer soundEffect.Close()

		done := make(chan bool, 1)

		go watchProcess(cmd.Process.Pid, done, &soundEffect, &format)

		cmd.Wait()

		done <- true
	} else {
		log.Fatal("Please specify a program")
	}
}
