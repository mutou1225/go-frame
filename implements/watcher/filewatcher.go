package watcher

import (
	"fmt"
	"github.com/radovskyb/watcher"
	"log"
	"time"
)

// 文件监控
// 返回Watcher
// 文件有变更通过 Watcher.Event 通知
// 错误信息通过 Watcher.Error 通知
// 结束信息通过 Watcher.Closed 通知
func FileWatcher(filePath string, d time.Duration) (*watcher.Watcher, error) {
	w := watcher.New()

	// SetMaxEvents to 1 to allow at most 1 event's to be received
	// on the Event channel per watching cycle.
	//
	// If SetMaxEvents is not set, the default is to send all events.
	w.SetMaxEvents(1)

	// Ops: Write
	w.FilterOps(watcher.Write)

	// Only files that match the regular expression during file listings
	// will be watched.
	//r := regexp.MustCompile("^abc$")
	//w.AddFilterHook(watcher.RegexFilterHook(r, false))

	// Watch this folder for changes.
	if err := w.Add(filePath); err != nil {
		return nil, err
	}

	// Watch test_folder recursively for changes.
	//if err := w.AddRecursive("../test_folder"); err != nil {
	//	return nil, err
	//}

	// Print a list of all of the files and folders currently
	// being watched and their paths.
	//for path, f := range w.WatchedFiles() {
	//	fmt.Printf("%s: %s\n", path, f.Name())
	//}

	// Trigger 2 events after watcher started.
	//go func() {
	//	w.Wait()
	//	w.TriggerEvent(watcher.Create, nil)
	//	w.TriggerEvent(watcher.Remove, nil)
	//}()

	// Start the watching process - it'll check for changes every d.
	go func() {
		if err := w.Start(d); err != nil {
			log.Println(fmt.Sprintf("FileWatcher Err: %s", err.Error()))
		}
	}()

	log.Println(fmt.Sprintf("FileWatcher File: %s Duration: %d", filePath, d))

	return w, nil
}
