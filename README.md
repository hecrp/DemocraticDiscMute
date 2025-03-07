# 🔇 DiscMuteBot

A Discord bot that allows users to democratically vote to mute other users in voice channels for a set period of time.

## ✨ Features

- Democratic voice channel moderation through voting
- Configurable vote threshold and duration
- Temporary muting that only affects voice channels (users can still type in text channels)
- Voice mute persists across channel changes
- Data persistence across bot restarts

## 🛠️ Requirements

- Go 1.16 or higher
- Discord bot token with proper permissions
- Bot requires administrator privileges on the server

## 🚀 Quick Setup

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/DiscMuteBot.git
   cd DiscMuteBot
   ```

2. Configure your bot token in `config.json`:
   ```json
   {
     "token": "YOUR_BOT_TOKEN"
   }
   ```

3. Install dependencies and build:
   ```
   go mod tidy
   go build -o DiscMuteBot bot/main.go
   ```

4. Generate an invitation link:
   ```
   cd utils
   go run generar_invitacion.go
   ```

5. Run the bot:
   ```
   ./DiscMuteBot
   ```

## 📝 Commands

- `!calladito @user` - Vote to mute the mentioned user in voice channels
- `!calladitoinfo` - Show all users with active votes
- `!calladitoinfo @user` - Show votes for a specific user
- `!calladitostatus` - Show mute system configuration
- `!calladitoping` - Check if the bot is active
- `!calladitodebug` - Show detailed information about the bot
- `!calladitoservidores` - Show servers where the bot is present
- `!calladitohelp` - Show available commands

## ⚙️ Advanced Configuration

You can modify these values in the `main.go` file:

- `VOTE_DURATION` - Duration of votes (default: 10 minutes)
- `MUTE_DURATION` - Duration of voice muting (default: 5 minutes)
- `VOTES_NEEDED` - Number of votes needed to mute a user (default: 5 votes/users)

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome. Please open an issue or a pull request to suggest changes or improvements. 