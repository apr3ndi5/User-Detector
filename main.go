package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type TTS struct {
	AllUsers string `json:"allusers"`
	UserList string `json:"userlist"`
}

// User type struct
type User struct {
	ID    string `json:"id"`
	Audio string `json:"audio"`
}

// User Config struct
var Config struct {
	Token     string `json:"token"`
	GuildID   string `json:"guild_id"`
	TTSConfig TTS    `json:"tts"`
	Audio     string `json:"audio"`
	Users     []User `json:"users"`
}

// Guild State
type GuildState struct {
	OldState *discordgo.Guild
}

var State GuildState

func main() {

	// Load Config File
	fmt.Println("Loading Config File")
	LoadConfig("./config.json")

	// Create a new Discord session using the provided bot token.
	fmt.Println("Creating Discord Session")
	Bot := InitBot(Config.Token)

	// Register the messageCreate func as a callback for VoiceStates events.
	Bot.AddHandler(VoiceStates)

	// In this example, we only care about receiving message events.
	Bot.Identify.Intents = discordgo.IntentsGuildVoiceStates

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	Bot.Close()

}

// Function to create a new Discord session
func InitBot(Token string) *discordgo.Session {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New(Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
	}
	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
	}
	return dg
}

// Function to load json file and decode it
func LoadConfig(path string) {
	// Load the config file
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("error opening config file,", err)
	}
	json.NewDecoder(file).Decode(&Config)
	defer file.Close()
}

// Function to run mp3 file
func RunSound(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	// Create a new beep.Decoder
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		fmt.Println(err)
	}
	defer streamer.Close()
	// Create a new beep.Player
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	// Play the stream
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

// Function to user speech in Powershell
func Speech(text string) {
	// Run the speech command
	cmd := exec.Command("Powershell.exe", "/Command", "Add-Type -AssemblyName", "System.speech;", "$speak = New-Object", "System.Speech.Synthesis.SpeechSynthesizer;", "$speak.Speak('", text, "')")
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

// VoiceStates is a function that will be called every time a new VoiceState is
func VoiceStates(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// Check if the user is in a voice channel
	for x := 0; x < len(Config.Users); x++ {
		// Checl of current GuildID equal to Config.GuildID
		if v.GuildID == Config.GuildID {
			// Save the old state
			State.OldState, _ = s.State.Guild(v.GuildID)
			for _, key := range State.OldState.VoiceStates {
				if key.UserID != v.UserID {
					continue
				}

				switch {
				case Config.TTSConfig.AllUsers == "true":
					name, err := s.User(v.UserID)
					if err != nil {
						fmt.Println(err)
					}
					Speech(name.Username + " join")
					if v.UserID == Config.Users[x].ID && v.UserID == s.State.User.ID {
						RunSound(Config.Audio)
					}
					return
				case Config.TTSConfig.UserList == "true":
					if v.UserID == Config.Users[x].ID && v.UserID == s.State.User.ID {
						name, err := s.User(v.UserID)
						if err != nil {
							fmt.Println(err)
						}
						Speech(name.Username + " join")
						RunSound(Config.Users[x].Audio)
					}
					return
				}
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
	time.Sleep(time.Second * 1)
}
