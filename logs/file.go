package logs

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"liadmire/sys"
)

type Logger struct {
	sync.RWMutex
	FileName      string    `json:"filename"`
	fileWriter    *os.File  `json:"-"`
	consoleWriter io.Writer `json:"-"`
	Level         LevelType `json:"leveltype"`
	MaxSize       int64     `json:"maxsize"`
	currentSize   int64     `json:"-"`
	Rotate        bool      `json:"rotate"`
	Perm          string    `json:"perm"`
	RotatePerm    string    `json:"rotateperm"`
}

func newLogger() *Logger {
	l := &Logger{
		FileName:      fmt.Sprintf("%s.log", sys.SelfNameWithoutExt()),
		consoleWriter: os.Stdout,
		Level:         DEBUG,
		MaxSize:       8192 * 8,
		Rotate:        true,
		Perm:          "0660",
		RotatePerm:    "0440",
	}

	l.startLogger()

	return l
}

func (l *Logger) doRotate() error {
	fName := ""
	// srcFile := ""
	// dstFile := ""
	_, err := os.Lstat(l.FileName)
	if err != nil {
		goto RESTART_LOGGER
	}

	if l.MaxSize > 0 && l.currentSize >= l.MaxSize {
		t := time.Now()
		fName = fmt.Sprintf("%s%d.log", t.Format("20060102150405"), t.UnixNano())
		_, err = os.Lstat(fName)
		if err != nil {
			fmt.Println("fName ...: ", fName, err.Error())
		}
	}

	if err == nil {
		return fmt.Errorf("Rotate: %s", l.FileName)
	}

	err = l.fileWriter.Close()
	if err != nil {
		fmt.Println("error:------> ", err.Error())
	}

	// srcFile = path.Join(sys.SelfDir(), l.FileName)
	// dstFile = path.Join(sys.SelfDir(), fName)
	//fmt.Printf("srcFile: %s, dstFile: %s\n", srcFile, dstFile)
	err = os.Rename(l.FileName, fName)
	if err != nil {
		fmt.Println("os.Rename: ", err.Error())
		goto RESTART_LOGGER
	}

RESTART_LOGGER:
	l.startLogger()

	return nil
}

func (l *Logger) writeMsg(level LevelType, msg string, v ...interface{}) error {
	l.Lock()
	defer l.Unlock()

	if l.Rotate {
		l.doRotate()
	}

	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}

	when := time.Now().Format("2006-01-02 15:04:05")
	prefix := levelPrefix[level]

	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}

	_, fileName := path.Split(file)

	msg = fmt.Sprintf("%s  %s [ %s : %d ] %s\n", prefix, when, fileName, line, msg)

	var err error

	if level == DEBUG {
		_, err = l.consoleWriter.Write([]byte(msg))
	} else {
		_, err = l.fileWriter.Write([]byte(msg))
		if err == nil {
			l.currentSize += int64(len(msg))
		}
	}

	fmt.Println("currentSize: ", l.currentSize)

	return err
}

func (l *Logger) startLogger() error {

	perm, err := strconv.ParseInt(l.Perm, 8, 64)
	if err != nil {
		return err
	}

	// filePath := path.Dir(l.FileName)
	// fmt.Println("startLogger: ", filePath)

	// os.MkdirAll(filePath, os.FileMode(perm))

	fd, err := os.OpenFile(l.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err != nil {
		return err
	} else {
		os.Chmod(l.FileName, os.FileMode(perm))
	}

	l.fileWriter = fd

	fInfo, err := fd.Stat()
	if err != nil {
		return err
	}

	l.currentSize = fInfo.Size()
	return nil
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.Level > DEBUG {
		return
	}

	l.writeMsg(DEBUG, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.Level > INFO {
		return
	}
	l.writeMsg(INFO, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.Level > WARN {
		return
	}
	l.writeMsg(WARN, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.Level > ERROR {
		return
	}
	l.writeMsg(ERROR, format, v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.Level > FATAL {
		return
	}
	l.writeMsg(FATAL, format, v...)
}

func (l *Logger) SetLevel(lt LevelType) {
	l.Level = lt
}

func (l *Logger) GetLevel() LevelType {
	return l.Level
}
