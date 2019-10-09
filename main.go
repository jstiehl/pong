package main

import (
  "fmt"
  "bufio"
  "os"
  "os/exec"
  "log"
  "time"
  "encoding/json"
)

type Ball struct {
  row int
  col int
  vx string
  vy string
}

type Paddle struct {
  row int
  col int
}

var ball Ball
var paddle Paddle
var lives int

// Config holds the emoji configuration
type Config struct {
  Ball     string `json:"ball"`
  Ghost    string `json:"ghost"`
  Wall     string `json:"wall"`
  Dot      string `json:"dot"`
  Pill     string `json:"pill"`
  Death    string `json:"death"`
  Space    string `json:"space"`
  UseEmoji bool   `json:"use_emoji"`
}

var cfg Config

func loadConfig() error {
  f, err := os.Open("config.json")
  if err != nil {
    return err
  }
  defer f.Close()

  decoder := json.NewDecoder(f)
  err = decoder.Decode(&cfg)
  if err != nil {
    return err
  }

  return nil
}

func loadArena() error {
  lives = 1
  f, err := os.Open("battle_field_1.txt")
  if err != nil {
    return err
  }
  defer f.Close()

  scanner := bufio.NewScanner(f)

  for scanner.Scan() {
    line := scanner.Text()
    arena = append(arena, line)
  }

  for row, line := range arena {
    for col, char := range line {
      switch char {
      case 'B':
        ball = Ball{row, col, "RIGHT", "DOWN"}
      case 'P':
        paddle = Paddle{row, col}
      }
    }
  }

  return nil
}

var arena []string

func clearScreen() {
  fmt.Printf("\x1b[2J")
  moveCursor(0, 0)
}

func moveCursor(row, col int) {
  if cfg.UseEmoji {
    fmt.Printf("\x1b[%d;%df", row+1, col*2+1)
  } else {
    fmt.Printf("\x1b[%d;%df", row+1, col+1)
  }
}

func makeMove(oldRow, oldCol int, dirx, diry string) (newRow, newCol int, newVx, newVy string) {
  newRow, newCol = oldRow, oldCol
  newVx, newVy = dirx, diry

  switch diry {
  case "UP":
    newRow = newRow - 1
  case "DOWN":
    newRow = newRow + 1
  }

  switch dirx {
    case "RIGHT":
      newCol = newCol + 1
    case "LEFT":
      newCol = newCol - 1
  }

  if arena[newRow][newCol] == '#' {
    //handle boundaries
    if newRow == 0 {
      newVy = "DOWN"
    } else if newRow == len(arena)-1 {
      newVy = "UP"
    } 

    if newCol == 0 {
      newVx = "RIGHT"
    } else if newCol == len(arena[0]) - 1 {
      newVx = "LEFT"
    } 
    // if there are other "obstacles" other than the boundaries, this code handles those collisions
    // if newRow != 0 && newRow != len(arena)-1 && newCol != 0 && newCol != len(arena[0]) - 1 {
    //   if arena[newRow][newCol + 1] == '#' || arena[newRow][newCol - 1] == '#' {
    //     if diry == "UP" {
    //       newVy = "DOWN"
    //     } else {
    //       newVy = "UP"
    //     }
    //   }

    //   if arena[newRow + 1][newCol] == '#' || arena[newRow - 1][newCol] == '#' {
    //     if dirx == "LEFT" {
    //       newVx = "RIGHT"
    //     } else {
    //       newVx = "LEFT"
    //     }
    //   }
    // }

    newRow = oldRow
    newCol = oldCol
  }

  //handle paddle collisions
  if newCol == paddle.col && (newRow == paddle.row || newRow == paddle.row + 1 || newRow == paddle.row - 1) {
    if dirx == "LEFT" {
      newVx = "RIGHT"
    } else {
      newVx = "LEFT"
    }
    newRow = oldRow
    newCol = oldCol
  }

  return
}

func moveBall() {
  ball.row, ball.col, ball.vx, ball.vy = makeMove(ball.row, ball.col, ball.vx, ball.vy)
  if ball.col < 2 {
    lives = 0
  }
}

func movePaddle(dir string) {
  switch dir {
    case "UP":
      //conditional keeps paddle within arena boundaries
      if paddle.row-2 != 0 {
        paddle.row = paddle.row - 1
      }
    case "DOWN":
      //conditional keeps paddle within arena boundaries
      if paddle.row + 2 != len(arena)-1 {
        paddle.row = paddle.row + 1
      }
  }
}

func readInput() (string, error) {
  buffer := make([]byte, 100)

  cnt, err := os.Stdin.Read(buffer)
  if err != nil {
    return "", err
  }

  if cnt == 1 && buffer[0] == 0x1b {
    return "ESC", nil
  } else if cnt >= 3 {
    if buffer[0] == 0x1b && buffer[1] == '[' {
      switch buffer[2] {
      case 'A':
        return "UP", nil
      case 'B':
        return "DOWN", nil
      }
    }
  }

  return "", nil
}

func printScreen() {
  clearScreen()
  for _, line := range arena {
    for _, chr := range line {
      switch chr {
      case '#':
        fmt.Printf(cfg.Wall)
      case 'G':
        fmt.Printf(cfg.Ghost)
      default:
        fmt.Printf(cfg.Space)
      }
    }
    fmt.Printf("\n")
  }

  //draw ball
  moveCursor(ball.row, ball.col)
  if lives == 0 {
    fmt.Printf(cfg.Death)
  } else {
    fmt.Printf(cfg.Ball)
  }

  //draw "paddle"
  moveCursor(paddle.row, paddle.col)
  fmt.Printf(cfg.Dot)
  moveCursor(paddle.row+1, paddle.col)
  fmt.Printf(cfg.Dot)
  moveCursor(paddle.row-1, paddle.col)
  fmt.Printf(cfg.Dot)
  //get cursor off arena
  moveCursor(len(arena), 0)
}

func init() {
  cbTerm := exec.Command("/bin/stty", "cbreak", "-echo")
  cbTerm.Stdin = os.Stdin

  err := cbTerm.Run()
  if err != nil {
    log.Fatalf("Unable to activate cbreak mode terminal: %v\n", err)
  }
}

func cleanup() {
  cookedTerm := exec.Command("/bin/stty", "-cbreak", "echo")
  cookedTerm.Stdin = os.Stdin

  err := cookedTerm.Run()
  if err != nil {
    log.Fatalf("Unable to activate cooked mode terminal: %v\n", err)
  }
}

func main() {
  defer cleanup()
  err := loadArena()

  err = loadConfig()
  if err != nil {
    log.Printf("Error loading configuration: %v\n", err)
    return
  }

  // process input (async)
  input := make(chan string)
  go func(ch chan<- string) {
    for {
      input, err := readInput()
      if err != nil {
        log.Printf("Error reading input: %v", err)
        ch <- "ESC"
      }
      ch <- input
    }
  }(input)

  for {
    // process movement
    select {
      case inp := <-input:
        if inp == "ESC" {
          lives = 0
        }
        if inp == "UP" || inp == "DOWN" {
          movePaddle(inp)
        }
      default:
    }
    moveBall()
    printScreen()
        // check game over
    if lives == 0 {
      break
    }

    // repeat
    time.Sleep(200 * time.Millisecond)
  }
}