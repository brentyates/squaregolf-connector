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

5. **🔴 Startup & Discovery**
   - Look for `Main()` or entry point
   - Check for command-line argument parsing
   - Find configuration file paths
   - Identify expected executable name/location

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

#### Strategy A: Standalone Mode (Independent Process)
Run our connector as a separate process, connect directly to GSPro:
- User manually starts our connector
- Connector connects to GSPro's internal port
- GSPro Connect is not needed at all
- **Advantage**: Clean separation, no GSPro modifications
- **Disadvantage**: User manages two processes

#### Strategy B: Drop-In Replacement (GSPro Launches Us)
Replace GSPro Connect executable with our connector:

**Option B1: Rename & Replace**
```bash
# Backup original GSPro Connect
mv GSProConnect.exe GSProConnect.exe.bak

# Build our connector as Windows executable
GOOS=windows GOARCH=amd64 go build -o GSProConnect.exe

# GSPro will now launch our connector instead
```

**Option B2: Config File Modification**
Find GSPro's config file (Unity's PlayerPrefs, XML, JSON, etc.) and change the connector path:
```bash
# Find config files in GSPro directory
find ~/.wine/drive_c/Program\ Files/GSPro -name "*.xml" -o -name "*.json" -o -name "*.config"

# Look for GSProConnect path references
grep -r "GSProConnect" ~/.wine/drive_c/Program\ Files/GSPro
```

**Files to modify:**
- `internal/core/gspro/gspro_direct.go` - Implement direct protocol
- `main.go` - Add Windows build target
- Build scripts for cross-compilation

#### Strategy C: Hybrid Mode (Already Implemented)
Toggle between modes in our connector:
- Default: Connect to GSPro Connect (port 921)
- Direct: Connect to GSPro directly (port TBD)
- **Advantage**: Flexibility, both modes available
- **Disadvantage**: Still need to manage processes

## Questions to Answer

### Critical Information Needed:
1. **Port Number**: What port does GSPro listen on?
2. **Protocol**: JSON like GSPro Connect, or different?
3. **Handshake**: Does GSPro require registration/authentication?
4. **Message Format**: Same as GSPro Connect API, or different?
5. **Keep-Alive**: Does GSPro expect periodic heartbeat?
6. **🔴 DISCOVERY MECHANISM**: How does GSPro find/launch GSPro Connect?
   - Does GSPro launch the connector executable itself?
   - Is there a config file (XML/JSON) with the connector path?
   - Does it look for a specific executable name/location?
   - What command-line arguments does GSPro pass to the connector?

### Decision Points:
1. Should this be:
   - [ ] Toggle in UI (Connect vs Direct)
   - [ ] Auto-detection with fallback
   - [ ] Command-line flag
   - [ ] Separate binary

2. Windows-only or cross-platform?
   - If protocol is pure TCP: works on Linux/Mac
   - If Windows-specific IPC: needs Wine/Windows

## Investigating the Discovery Mechanism

### Finding How GSPro Launches GSPro Connect

**1. Search for Config Files:**
```bash
# Unity typically stores config in these places:
# Windows: C:\Users\<user>\AppData\LocalLow\<company>\<product>
# Wine equivalent:
find ~/.wine/drive_c/users/*/AppData/LocalLow -name "*GSPro*" -type f

# Also check GSPro installation directory:
find ~/.wine/drive_c/Program\ Files*/GSPro* -name "*.xml" -o -name "*.json" -o -name "*.ini" -o -name "*.config"

# Look for any files mentioning "Connect":
grep -r "Connect" ~/.wine/drive_c/Program\ Files*/GSPro* 2>/dev/null | grep -i "\.exe\|path\|launch"
```

**2. Monitor GSPro Process:**
```bash
# While GSPro is running, check what processes it spawned:
ps aux | grep -i gspro

# Check process tree:
pstree -p | grep -i gspro

# If GSPro launches GSPro Connect, you'll see the parent-child relationship
```

**3. Check for Registry Keys (Wine):**
```bash
# Unity apps sometimes store paths in registry
# Wine registry: ~/.wine/user.reg or system.reg
grep -i "gspro" ~/.wine/*.reg
grep -i "connect" ~/.wine/*.reg
```

**4. Decompile GSPro Itself (If Needed):**
If GSPro is also .NET/Unity, we can decompile it to find:
- How it discovers connectors
- What it looks for (executable name, registry key, config file)
- Command-line arguments it passes

### Expected Discovery Patterns

**Pattern 1: Fixed Executable Name/Location**
```
GSPro expects: C:\Program Files\GSPro\GSProConnect.exe
Solution: Replace with our exe named GSProConnect.exe
```

**Pattern 2: Config File Path**
```xml
<!-- Unity PlayerPrefs or custom config -->
<GSPro>
  <ConnectorPath>C:\Program Files\GSPro\GSProConnect.exe</ConnectorPath>
</GSPro>

Solution: Update config to point to our exe
```

**Pattern 3: Registry Key**
```
HKEY_LOCAL_MACHINE\SOFTWARE\GSPro\ConnectorPath
Solution: Update Wine registry
```

**Pattern 4: Plugin Directory Scan**
```
GSPro scans: C:\Program Files\GSPro\Plugins\
Loads any exe matching pattern: *Connect*.exe
Solution: Place our exe in plugins directory
```

### Build for Windows

Once we know the discovery mechanism, build for Windows:

```bash
# Cross-compile for Windows from Linux
GOOS=windows GOARCH=amd64 go build -o squaregolf-connector.exe

# If we need to match GSPro Connect's name:
GOOS=windows GOARCH=amd64 go build -o GSProConnect.exe

# Copy to Wine GSPro directory:
cp GSProConnect.exe ~/.wine/drive_c/Program\ Files/GSPro/
```

## Next Steps

1. [ ] Locate GSPro Connect executable
2. [ ] **🔴 Investigate how GSPro discovers/launches GSPro Connect**
3. [ ] Decompile GSPro Connect with dnSpy/ILSpy
4. [ ] Extract protocol details
5. [ ] Validate with network capture
6. [ ] Document protocol
7. [ ] **🔴 Determine deployment strategy (replace exe vs config vs standalone)**
8. [ ] Implement direct mode in connector
9. [ ] Build Windows executable
10. [ ] Test with actual GSPro
11. [ ] Create user documentation

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
