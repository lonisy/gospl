package library

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

func Watcher(sig syscall.Signal) {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath) // 解析可能的软链接
	if err != nil {
		log.Fatalf("Failed to resolve symlink: %v", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()
	dir := filepath.Dir(execPath)
	err = watcher.Add(dir)
	if err != nil {
		log.Fatalf("Failed to watch directory: %v", err)
	}
	fmt.Println("Watching executable:", execPath)
	fmt.Println("Watching directory:", dir)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			fmt.Println("select-event:", event.Name)
			fmt.Println("select-event-data:", event.String())
			if event.Name == execPath && event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				syscall.Kill(os.Getpid(), sig)
				return
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}
