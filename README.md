# Chronos (Software Archaeology Simulator)

Chronos is a desktop application designed to reverse-engineer believable, auditable Git histories for completed or sparse software projects. Unlike simple "commit spammers," Chronos uses a "Shadow Git" architecture and Reverse-Diff methodology to mathematically deconstruct a final codebase and rebuild it as a methodical, months-long enterprise development cycle.

## 📄 Product Requirements Document (PRD): Chronos
**Version:** 1.0
**Status:** Approved for Development

### 1. Executive Summary
It transforms a single "initial commit" into a compelling, flawed, and entirely believable project mythology.

### 2. Target Audience & Use Cases
*   **Freelancers & Agencies:** Need to prove methodical development cycles and sprint histories to clients for completed deliverables.
*   **Solo Developers:** Want to legitimize weekend hackathon projects or personal portfolios to look like sustained, professional efforts.
*   **Security Researchers:** Need to generate realistic-looking repository histories for honeypots or testing environments.

### 3. Technical Architecture (Zero JavaScript)
*   **Frontend (UI):** Flutter (Dart) Desktop Application.
*   **Backend (Engine):** Go (Golang) Compiled Binary.
*   **Communication (IPC):** JSON over STDIN/STDOUT (NDJSON logs).

### 4. Core Engine Mechanics: The "Shadow Git" Method
1.  **Snapshot:** Reads the user's target directory state.
2.  **Shadow Init:** A temporary repo is created.
3.  **Timeline Generation:** Dummy files and configs are committed with backdated timestamps.
4.  **The Overlay Step:** Final project files are copied and committed.
5.  **Replacement:** Replaces the user's `.git` folder with the Shadow Repo's `.git` folder.

### 5. Engine Configuration Modes
*   **Deterministic:** Pure Go logic and predefined templates.
*   **API-Driven:** OpenAI-compatible REST API calls (Pollinations, Ollama, OpenAI, etc.).

### 6. Realism Injectors
*   AST Code Phasing
*   Dependency Timeline Alignment
*   Human Error Injection
*   Branch & PR Simulation
*   Chronobiometrics
*   Phantom Team

---

## 🛠 Build Instructions

### Prerequisites
- **Go** (1.24.3 or later)
- **Flutter** (3.41.2 or later)
- **Make** (optional, for automation)

### Platforms Supported
- **Windows**
- **Linux**

### Building from Source

#### Using Makefile (Linux/WSL/MacOS with Linux tools)
```bash
make all
```

#### Manual Build

**1. Build the Go Engine:**
```bash
cd engine
go build -o chronos-engine main.go
```
*On Windows:*
```bash
go build -o chronos-engine.exe main.go
```

**2. Build the Flutter UI:**
```bash
cd ui
flutter build linux # or windows
```

### Running the Application
Ensure the `chronos-engine` binary is in the same directory as the UI executable or set the `CHRONOS_ENGINE_PATH` environment variable.
