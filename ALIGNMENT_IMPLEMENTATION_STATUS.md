# Alignment Feature Implementation Status

## Completed (Backend Foundation)

###  1. Data Structures (`internal/core/parse_notifications.go`)
- ✅ Added `AlignmentData` struct with:
  - `RawData []string` - raw bluetooth bytes for debugging
  - `AimAngle float64` - degrees left (negative) or right (positive)
  - `IsAligned bool` - whether within tolerance threshold (±2°)
- ✅ Added `ParseAlignmentData()` function with:
  - Placeholder parsing logic (assumes bytes 3-4 contain int16 angle / 100.0)
  - Alignment threshold calculation
  - Error handling

### 2. State Management (`internal/core/state_manager.go`)
- ✅ Added `IsAligning bool` to AppState - tracks whether alignment mode is active in UI
- ✅ Added `AlignmentAngle float64` to AppState - current aim angle
- ✅ Added getter/setter methods:
  - `GetIsAligning()` / `SetIsAligning()`
  - `GetAlignmentAngle()` / `SetAlignmentAngle()`
- ✅ Added callback registration:
  - `RegisterIsAligningCallback()`
  - `RegisterAlignmentAngleCallback()`

### 3. Notification Handling (`internal/core/launch_monitor.go`)
- ✅ Added alignment notification detection (format: `11 82`)
- ✅ Added `HandleAlignmentNotification()` method
- ✅ Integrated into `NotificationHandler()` routing
- ✅ Added debug logging for received angles

## TODO - Requires Device Testing

### 1. Bluetooth Protocol Discovery
**Status:** 🔴 BLOCKED - Need physical device

**Tasks:**
- [ ] Connect official Square Golf app to device
- [ ] Enable Bluetooth HCI Snoop Log on Android
- [ ] Capture traffic while using alignment feature
- [ ] Identify:
  - Correct BLE characteristic UUID for alignment data
  - Exact byte format (currently guessing `11 82` header)
  - Byte positions for aim angle
  - Data type and scaling factor
  - Update frequency

**Current Assumptions (need verification):**
```go
// Assumed format in ParseAlignmentData:
// - Header: bytes [0-1] = "11 82"
// - Angle: bytes [3-4] = int16 (little-endian) / 100.0
// - Similar to other angle fields in protocol
```

### 2. Frontend Implementation
**Status:** 🟡 IN PROGRESS

**Web UI Files to Update:**
- [ ] `web/static/js/app.js`:
  - Add alignment data handling in WebSocket message handler
  - Create `updateAlignmentData(data)` method
  - Handle `isAligning` and `alignmentAngle` state

- [ ] `web/index.html` - Alignment screen section:
  - Add numeric angle display ("Aimed 12.3° right")
  - Add visual compass indicator (SVG/Canvas)
  - Add Start/Stop Alignment buttons
  - Add alignment status indicator

- [ ] `web/static/css/style.css`:
  - Style compass visual indicator
  - Add color coding (green = aligned, yellow/red = off-target)
  - Responsive layout

**Compass Design Ideas:**
```
    ┌───────────────┐
    │   ALIGNMENT   │
    ├───────────────┤
    │               │
    │    ←  ●  →    │  Visual indicator
    │               │
    │  Aimed:       │
    │  12.3° right  │  Numeric display
    │               │
    │ [Start Align] │  Button
    └───────────────┘
```

### 3. Backend Commands
**Status:** 🔴 NOT STARTED

**Need to implement:**
- [ ] `StartAlignment()` method in LaunchMonitor
  - Send command to device to start streaming alignment data
  - Set `IsAligning = true` in state

- [ ] `StopAlignment()` method in LaunchMonitor
  - Send command to stop alignment data stream
  - Set `IsAligning = false` in state

- [ ] Discover command format from Bluetooth capture:
  ```go
  // Example placeholder:
  func StartAlignmentCommand(sequence int) string {
      return fmt.Sprintf("118X%02x000000000000", sequence)
      // X = command ID (unknown, need to discover)
  }
  ```

### 4. WebSocket Integration
**Status:** 🔴 NOT STARTED - Need to locate web server code

**Tasks:**
- [ ] Find/create WebSocket server implementation
- [ ] Add `alignmentData` message type
- [ ] Broadcast alignment angle updates
- [ ] Broadcast `isAligning` state changes
- [ ] Include in initial state payload on connection

## Testing Checklist

### With Real Device:
1. [ ] Connect device via Bluetooth
2. [ ] Capture official app traffic during alignment
3. [ ] Identify BLE characteristic UUID
4. [ ] Verify byte format and positions
5. [ ] Update `ParseAlignmentData()` with correct format
6. [ ] Test alignment angle updates in real-time
7. [ ] Verify ±2° threshold works correctly
8. [ ] Test start/stop alignment commands

### UI Testing:
1. [ ] Verify compass visual updates smoothly
2. [ ] Verify numeric display shows correct values
3. [ ] Test left (negative) and right (positive) angles
4. [ ] Verify Start/Stop buttons toggle correctly
5. [ ] Test alignment works in all browsers (Chrome, Firefox, Safari)

## Known Limitations

1. **No Mock Data**: Currently no simulation mode for alignment feature
   - Cannot test UI without real device
   - Consider adding mock alignment data to simulator for development

2. **Update Frequency Unknown**: Don't know how often device sends alignment data
   - May need throttling if too frequent
   - May need interpolation if too slow

3. **Calibration**: Device has calibration offsets in firmware
   - `ACC_SENSOR_TILT = -1.827°`
   - `ACC_SENSOR_ROLL = 1.550°`
   - Unknown if these are applied device-side or need client-side application

## Files Modified

- `internal/core/parse_notifications.go` - AlignmentData struct and parsing
- `internal/core/state_manager.go` - State fields and callbacks
- `internal/core/launch_monitor.go` - Notification handling

## Files To Modify

- `web/static/js/app.js` - JavaScript handling
- `web/index.html` - UI elements
- `web/static/css/style.css` - Styling
- WebSocket server file (location TBD)
- `internal/core/commands.go` - Start/Stop alignment commands

## Next Steps

**Immediate (when device available):**
1. Capture Bluetooth traffic from official app
2. Document exact protocol format
3. Update parsing code with real format
4. Test backend with real data

**After Protocol Confirmed:**
1. Implement WebSocket broadcasting
2. Build frontend UI components
3. Add start/stop commands
4. End-to-end testing

---

**Last Updated:** 2025-10-31
**Branch:** `alignment`
**Status:** Backend foundation complete, waiting for device testing
