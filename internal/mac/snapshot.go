//go:build darwin

package mac

import "sort"

// Snapshot gathers displays, screens and every application's AX windows.
// The pid set is the union of NSWorkspace's regular applications and the
// CG window list's layer-0 owners, so hidden apps and other-Space windows
// are seen even though CGWindowList omits them (spike fact). Each app gets
// a messaging timeout so one hung process cannot stall the sweep.
func Snapshot(timeoutSeconds float32) SnapshotRaw {
	Load()
	snap := SnapshotRaw{CGWindows: WindowList()}
	snap.Displays = Displays()
	snap.Screens, snap.PrimaryH = Screens()

	apps := map[int64]AppRaw{}
	var pids []int64
	for _, a := range RunningApps() {
		apps[a.PID] = a
		pids = append(pids, a.PID)
	}
	for _, w := range snap.CGWindows {
		if w.Layer != 0 {
			continue
		}
		if _, ok := apps[w.OwnerPID]; !ok {
			apps[w.OwnerPID] = AppRaw{PID: w.OwnerPID, Bundle: BundleID(w.OwnerPID), Name: w.OwnerName}
			pids = append(pids, w.OwnerPID)
		}
	}
	sort.Slice(pids, func(i, j int) bool { return pids[i] < pids[j] })

	for _, pid := range pids {
		aw := AppWindows{App: apps[pid], AppEl: AXApp(int32(pid), timeoutSeconds)}
		aw.Hidden, _ = aw.AppEl.Bool("AXHidden")
		wins, err := aw.AppEl.Windows()
		if err != nil {
			aw.Err = err.Error()
			snap.Apps = append(snap.Apps, aw)
			continue
		}
		for _, w := range wins {
			raw := WindowRaw{
				El:      w,
				Title:   w.String("AXTitle"),
				Role:    w.String("AXRole"),
				Subrole: w.String("AXSubrole"),
			}
			raw.Minimized, _ = w.Bool("AXMinimized")
			raw.Pos, _ = w.Point("AXPosition")
			raw.Size, _ = w.Size("AXSize")
			aw.Windows = append(aw.Windows, raw)
		}
		snap.Apps = append(snap.Apps, aw)
	}
	return snap
}
