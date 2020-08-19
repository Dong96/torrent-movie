package encode

import (
	"bytes"
	"encode-service/model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type VideoInfo struct {
	FormatCont Format   `json:"format"`
	Streams    []Stream `json:"streams"`
}

type Format struct {
	Name     string `json:"filename"`
	Duration string `json:"duration"`
	Bitrate  string `json:"bit_rate"`
	Size     string `json:"size"`
}

type Stream struct {
	Index     int    `json:"index"`
	Type      string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Duration  string `json:"duration"`
	Bitrate   string `json:"bit_rate"`
	Channels  int    `json:"channels"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Tags      Tags   `json:"tags"`
}

type Tags struct {
	Bitrate  string `json:"BPS"`
	Duration string `json:"DURATION"`
}

func GetVideoInfo(videoPath string) (*VideoInfo, error) {
	script := "./info.sh"
	// path := filepath.FromSlash(videoPath)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(script, videoPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Cannot get videoinfo of %v, error: %v", filepath.Base(videoPath), err)
	}

	if stderr.String() != "" {
		return nil, fmt.Errorf("Stderr: %v", stderr.String())
	}

	vi := new(VideoInfo)
	if err := json.Unmarshal(stdout.Bytes(), vi); err != nil {
		return nil, fmt.Errorf("Cannot unmarshal videoinfo: %v", err)
	}

	// mili := (numArr[0]*60*60 + numArr[1]*60 + numArr[2]) * 1000

	return vi, nil
}

var taskMap encodeTask

func TaskMap() *encodeTask {
	return &taskMap
}

func init() {
	taskMap.task = make(map[string]*encode)
}

type encodeTask struct {
	mu   sync.RWMutex
	task map[string]*encode
}

func (et *encodeTask) add(e *encode) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.task[e.name] = e
}

func (et *encodeTask) get(key string) (*encode, bool) {
	et.mu.RLock()
	defer et.mu.RUnlock()
	v, ok := et.task[key]
	return v, ok
}

func (et *encodeTask) delete(key string) {
	et.mu.Lock()
	defer et.mu.Unlock()
	delete(et.task, key)
}

func (et *encodeTask) isEncode(key string) bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	v, ok := et.get(key)
	return ok && v.isEncode
}

func (et *encodeTask) Len() int {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return len(et.task)
}

type encode struct {
	name     string
	isEncode bool
	isDone   bool
	cmd      *exec.Cmd
}

type EncodeProgress struct {
	Name     string  `json:"name"`
	Progress float32 `json:"progress"`
	Status   string  `json:"status"`
}

func EncodeVideo(w http.ResponseWriter, videoPath, name string) error {
	if err := removeUndoneEncode(videoPath); err != nil {
		return err
	}
	if taskMap.isEncode(name) {
		return model.ClientError{nil, "Already Encoding!", http.StatusConflict}
	}
	script := "./encode.sh"
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	vi, err := GetVideoInfo(videoPath)
	if err != nil {
		return err
	}
	var audioIndex, channels, ab string
	var i, c int
	for _, stream := range vi.Streams {
		if c < stream.Channels {
			c = stream.Channels
			i = stream.Index
			ab = stream.Bitrate
		}
	}
	// encode with 80% bitrate
	b, err := strconv.ParseInt(ab, 10, 64)
	if err != nil {
		return err
	}
	b /= 100 * 80
	audioIndex = strconv.Itoa(i)
	channels = strconv.Itoa(c)
	audioBitrate := strconv.FormatInt(b, 10)

	path := filepath.FromSlash(videoPath)
	cmd := exec.Command(script, path, audioIndex, channels, audioBitrate)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	e := &encode{name, false, false, cmd}
	taskMap.add(e)

	if err := cmd.Start(); err == nil {
		startTime := time.Now()
		e.isEncode = true
		go func() {
			defer taskMap.delete(name)
			if err := cmd.Wait(); err != nil {
				log.Printf(`Stop encode "%v" when not finished!`, e.name)
				e.isEncode = false
				return
			}
			total := time.Since(startTime)
			log.Printf(`Encode "%v" Completed! in %v`, e.name, total)
			e.isDone = true

			out := strings.TrimSpace(stdout.String())
			err := strings.TrimSpace(stderr.String())
			if out != "" {
				log.Println(out)
			}
			if err != "" {
				log.Println(err)
			}
		}()
	} else {
		taskMap.delete(e.name)
		return err
	}

	return nil
}

func removeUndoneEncode(videoPath string) error {
	path := filepath.Dir(videoPath)
	path = filepath.Join(path, "dash")
	return os.RemoveAll(path)
}

func StopEncode(w http.ResponseWriter, videoName string) error {
	if taskMap.isEncode(videoName) {
		taskMap.mu.RLock()
		defer taskMap.mu.RUnlock()
		e := taskMap.task[videoName]
		log.Println("Call Stop encode")
		if err := syscall.Kill(-e.cmd.Process.Pid, syscall.SIGINT); err != nil {
			return fmt.Errorf(`Failed to stop encode video "%v", error: %v`, videoName, err)
		}
		e.isEncode = false
	}
	return model.ClientError{nil, "Cannot find video or alreadt stopped!", http.StatusNotFound}
}

func GetEncodeProgress(path, name string) (*EncodeProgress, error) {
	vi, err := GetVideoInfo(path)
	if err != nil {
		return nil, err
	}
	duration, err := getMilisDuration(vi.FormatCont.Duration)
	if err != nil {
		return nil, err
	}

	encodeTime, status, err := GetEncodeTime(path)
	if err != nil {
		return nil, err
	}

	percent := float32(encodeTime) / float32(duration) * 100
	if percent > 99.9 {
		percent = 100
	}
	return &EncodeProgress{name, percent, status}, nil
}

func GetEncodeTime(videoPath string) (int, string, error) {
	path := filepath.Join(filepath.Dir(videoPath), "dash/block.txt")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`tail -n 13 "%v" | grep "out_time_us\|progress"`, path))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return 0, "", fmt.Errorf("%v: %v", err, stderr.String())
	}
	if stderr.String() != "" {
		return 0, "", fmt.Errorf(stderr.String())
	}

	output := strings.Split(stdout.String(), "\n")
	time := strings.Split(output[0], "=")[1]
	time = strings.TrimSpace(time)
	mili, err := strconv.Atoi(time)
	if err != nil {
		return 0, "", err
	}

	progress := strings.Split(output[1], "=")[1]
	progress = strings.TrimSuffix(progress, "\n")

	mili /= 1000

	return mili, progress, nil
}

func getMilisDuration(d string) (int, error) {
	arr := strings.Split(d, ":")
	timeArr := make([]float64, len(arr))
	for i := range arr {
		t, err := strconv.ParseFloat(arr[i], 32)
		if err != nil {
			return -1, err
		}
		timeArr[i] = t
	}
	milis := int((timeArr[0]*60*60 + timeArr[1]*60 + timeArr[2]) * 1000)
	return milis, nil
}
