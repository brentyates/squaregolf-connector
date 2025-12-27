# GSPro Direct Integration - Reverse Engineering Plan

## Goal
Reverse engineer the GSPro Connect ↔ GSPro protocol to enable direct integration, eliminating the need for GSPro Connect.

## Current Setup
- GSPro runs in Wine (Windows app)
- GSPro Connect runs in Wine (C# application, closed source)
- SquareGolf Connector → GSPro Connect (port 921, documented)
- GSPro Connect → GSPro (port unknown, protocol unknown)

## Reverse Engineering Approach

### Phase 1: Locate GSPro Connect Binary
1. Find GSPro Connect executable in Wine prefix
   ```bash
   find ~/.wine -name "*GSPro*Connect*.exe" -o -name "*Connect*.exe"
   ```
2. Identify .NET version (Framework vs Core)

### Phase 2: Decompile GSPro Connect
**Recommended Tool: dnSpy** (most powerful, has debugger)
- Download: https://github.com/dnSpy/dnSpy/releases
- Alternative: ILSpy (cross-platform .NET decompiler)

**What to Extract:**
1. **Port Number**
   - Search for: `TcpClient`, `Socket`, `Connect`, port numbers
   - Look for configuration/constants

2. **Protocol Format**
   - Find message serialization (JSON? Binary? Protobuf?)
   - Message structure/schemas
   - Encoding (UTF-8? ASCII?)

3. **Connection Flow**
   - Initial handshake
   - Authentication (if any)
   - Keep-alive mechanism
   - Disconnection handling

4. **Message Types**
   - Shot data format to GSPro
   - Ready/status messages from GSPro
   - Player information flow
   - Error handling

### Phase 3: Network Capture (Validation)
Capture actual traffic to validate decompiled findings:

```bash
# Terminal 1: Start packet capture
sudo tcpdump -i lo -w ~/gspro-internal-capture.pcap -v

# Terminal 2: Start GSPro and GSPro Connect
# Terminal 3: Start SquareGolf Connector
# Perform test shots

# Stop capture, analyze in Wireshark
wireshark ~/gspro-internal-capture.pcap
```

**What to Validate:**
- Actual port number
- Message format matches decompiled code
- Handshake sequence
- Timing/keep-alive intervals

### Phase 4: Protocol Documentation
Document findings:
1. Port number
2. Message format (JSON/binary/other)
3. All message types with examples
4. Connection lifecycle
5. Error conditions

### Phase 5: Implementation Plan

#### Option A: Direct Mode Toggle
Add configuration option to connector:
```go
type GSProMode string

const (
    GSProModeConnect GSProMode = "connect" // Current: GSPro Connect
    GSProModeDirect  GSProMode = "direct"  // New: Direct GSPro
)
```

**Files to modify:**
- `internal/core/gspro/connection.go` - Add direct mode connection logic
- `internal/core/gspro/messages.go` - Add direct mode protocol (if different)
- `internal/config/config.go` - Add mode configuration
- `internal/web/server.go` - Add API for mode selection
- `web/static/js/services/GSProService.js` - Add UI toggle

#### Option B: Automatic Detection
Try direct connection first, fall back to GSPro Connect if it fails.

## Questions to Answer

### Critical Information Needed:
1. **Port Number**: What port does GSPro listen on?
2. **Protocol**: JSON like GSPro Connect, or different?
3. **Handshake**: Does GSPro require registration/authentication?
4. **Message Format**: Same as GSPro Connect API, or different?
5. **Keep-Alive**: Does GSPro expect periodic heartbeat?

### Decision Points:
1. Should this be:
   - [ ] Toggle in UI (Connect vs Direct)
   - [ ] Auto-detection with fallback
   - [ ] Command-line flag
   - [ ] Separate binary

2. Windows-only or cross-platform?
   - If protocol is pure TCP: works on Linux/Mac
   - If Windows-specific IPC: needs Wine/Windows

## Next Steps

1. [ ] Locate GSPro Connect executable
2. [ ] Decompile with dnSpy/ILSpy
3. [ ] Extract protocol details
4. [ ] Validate with network capture
5. [ ] Document protocol
6. [ ] Implement direct mode in connector
7. [ ] Test with actual GSPro
8. [ ] Create user documentation

## Tools Needed

- **dnSpy**: .NET decompiler/debugger
- **Wireshark**: Network protocol analyzer
- **tcpdump**: Packet capture
- **ILSpy** (alternative): Cross-platform .NET decompiler

## Risk Assessment

**Low Risk:**
- C# decompiles excellently (almost original source)
- We already understand one side (GSPro Connect API)
- Network protocol is likely TCP (reversible)

**Challenges:**
- Protocol might be binary (harder than JSON)
- GSPro might do version checking
- Undocumented quirks/edge cases

**Legal:** Personal use, interoperability, no redistribution of GSPro code = legally sound
