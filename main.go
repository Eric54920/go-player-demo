package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// 全局变量
var (
	ctrl     *beep.Ctrl
	volume   *effects.Volume
	speedy   *beep.Resampler
	streamer beep.StreamSeekCloser
	format   beep.Format
)

func Play(filePath string) {
	var err error
	// 读取文件流
	fileStream, _ := os.Open(filePath)

	defer fileStream.Close()

	// 获取文件扩展名并转换为小写
	ext := strings.ToLower(filepath.Ext(filePath))

	// 根据扩展名选择解码器
	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(fileStream)
	case ".wav":
		streamer, format, err = wav.Decode(fileStream)
	case ".flac":
		streamer, format, err = flac.Decode(fileStream)
	default:
		log.Fatalf("Unsupported file type: %s", ext)
	}

	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	// 初始化扬声器
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	// 创建控制器、音量控制和播放速度控制
	loop := beep.Loop(-1, streamer)
	ctrl = &beep.Ctrl{Streamer: loop, Paused: false}
	volume = &effects.Volume{Streamer: ctrl, Base: 2, Volume: 0, Silent: false}
	speedy = beep.ResampleRatio(4, 1, volume)
	speaker.Play(speedy)

	for {
		printInfo()
		time.Sleep(1 * time.Second)
	}
}

// togglePause 切换播放/暂停状态
func TogglePause() {
	speaker.Lock()
	ctrl.Paused = !ctrl.Paused
	speaker.Unlock()
}

// increaseVolume 增加音量
func IncreaseVolume(amount float64) {
	speaker.Lock()
	volume.Volume += amount
	speaker.Unlock()
}

// decreaseVolume 减少音量
func DecreaseVolume(amount float64) {
	speaker.Lock()
	volume.Volume -= amount
	speaker.Unlock()
}

// increaseSpeed 增加播放速度
func IncreaseSpeed(ratio float64) {
	speaker.Lock()
	speedy.SetRatio(speedy.Ratio() + ratio)
	speaker.Unlock()
}

// decreaseSpeed 减少播放速度
func DecreaseSpeed(ratio float64) {
	speaker.Lock()
	speedy.SetRatio(speedy.Ratio() - ratio)
	speaker.Unlock()
}

// seekByDuration 快进或倒退指定的时间
func SeekByDuration(duration time.Duration) {
	speaker.Lock()
	current := streamer.Position()
	newPosition := format.SampleRate.N(format.SampleRate.D(current).Round(time.Second) + duration)
	if newPosition < 0 {
		newPosition = 0
	} else if newPosition > streamer.Len() {
		newPosition = streamer.Len()
	}
	streamer.Seek(newPosition)
	speaker.Unlock()
}

func printInfo() {
	fmt.Printf("Position: %v\n", format.SampleRate.D(streamer.Position()).Round(time.Second))
	fmt.Printf("Volume: %v\n", volume.Volume)
	fmt.Printf("Speed: %v\n", speedy.Ratio())
}

func main() {
	filePath := "1.mp3"
	Play(filePath)
}
