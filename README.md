# Pokémon Clash

Pokémon Clash is a terminal-based online multiplayer Pokémon battle simulator written in Go. It uses a server-client architecture over WebSockets, featuring a real-time combat engine and a TUI made using bubbletea.

---

## TODO

* Fixing the issue with the TUI breaking for the first mon and move for some reason
* Add switch support
* Add a more immersive battle menu with sprites maybe? to look more like the GBA battle screen

---

## Setup & Running Guide

Follow these steps to run the game locally:

### Prerequisites
Make sure you have [Go](https://go.dev) installed on your system.

### Step 1: Create the Teams 
Create a team json file in `data/team<teamname>.json`
It should have the format:
```bash
[
  { "mon": "<name of the pokemon>", "moves": ["<move1>", "<move2>", "<move3>", "<move4>"] },
  ...and so on with the other 5 mons.
]
```

### Step 2: Download Datasets
Initialize the application by running the data fetcher script. This will download `pokedex.json` and `moves.json` into a local `data/` folder:
```bash
go run cmd/scripts/fetcher.go
```

### Step 3: Spin Up the Server
Start the backend WebSocket matchmaker and battle engine:
```bash
go run cmd/server/main.go
```
The server will start listening on port `:8080`.

### Step 4: Run the Game Client
Launch a terminal client session (open a second terminal or run multiple sessions to test matchmaking):
```bash
go run cmd/client/main.go data/team<teamname>.json 
```
Once two clients connect, they will be matched immediately and the battle will begin.

---

## Gameplay & Controls

* **Matchmaking**: Connecting to the server queues you automatically. Each player starts with a team of **Gengar**, **Venusaur**, and **Charizard**. This can be modified by modifying the data in ```cmd/client/main.go```.
* **Battle Layout**:
  * **Left Panel**: Scrolling log detailing battle events (e.g., moves used, damage dealt, and fainted Pokémon).
  * **Right Panel**: Listing of your current active/benched team and key bindings.
* **Controls**:
  * `1` : Select and execute Move 1
  * `2` : Select and execute Move 2
  * `3` : Select and execute Move 3
  * `4` : Select and execute Move 4
  * `q` / `Ctrl+C` : Gracefully disconnect and exit the program
