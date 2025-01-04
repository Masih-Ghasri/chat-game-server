# Game Chat System

This project implements a chat system for a multiplayer game where players can communicate within specific game maps/regions. Players can join the game, switch between maps, and send messages to other players in the same map.

---

## **Features**
- **Player Management**: Add players to the game with unique names.
- **Map Management**: Create multiple game maps/regions.
- **Chat System**: Players can send and receive messages within their current map.
- **Concurrency Support**: Handles multiple players and messages concurrently using Go's goroutines and channels.
- **Synchronization**: Uses mutexes to prevent race conditions when accessing shared data.

---

## **How It Works**
1. **Game Initialization**: Create a new game with a list of map IDs.
2. **Player Connection**: Players can join the game with a unique name.
3. **Map Switching**: Players can move between different maps.
4. **Messaging**: Players can send messages to their current map, and all players in that map will receive the message.

---

## **Code Structure**
- **`Player`**: Represents a player in the game.
  - `name`: The player's name.
  - `mapID`: The ID of the map the player is currently in.
  - `channel`: A channel for receiving messages.
  - `game`: A reference to the game the player belongs to.

- **`Map`**: Represents a game map/region.
  - `id`: The map's unique ID.
  - `players`: A list of players currently in the map.
  - `messages`: A channel for broadcasting messages to players in the map.
  - `mu`: A mutex for synchronizing access to the map's players.

- **`Game`**: Represents the main game instance.
  - `players`: A list of all players in the game.
  - `maps`: A list of all maps in the game.
  - `mu`: A mutex for synchronizing access to the game's players and maps.

---

## **Usage**
### 1. Create a New Game
```go
game, err := NewGame([]int{1, 2, 3})
if err != nil {
    log.Fatal("Error creating game:", err)
}