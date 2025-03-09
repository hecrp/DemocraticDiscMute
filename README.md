# 🔇 DiscMuteBot

A Discord bot that allows users to democratically vote to mute other users in voice channels for a set period of time.

## ✨ Features

- Democratic voice channel moderation through voting
- Configurable vote threshold and duration
- Temporary muting that only affects voice channels (users can still type in text channels)
- Voice mute persists across channel changes
- Data persistence across bot restarts
- Admin command to clear votes and unmute users
- CSV logging system for all bot activities

## 🛠️ Requirements

- Go 1.22 or higher
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

4. Run the bot:
   ```
   ./DiscMuteBot
   ```

### Docker Setup

You can run the bot using Docker with these simple steps:

1. Build the Docker image:
   ```bash
   docker build -t discmutebot .
   ```

2. Run the container:
   ```bash
   docker run -d --name mute-bot \
     -v $(pwd)/config.json:/app/config.json \
     -v $(pwd)/logs:/app/logs \
     discmutebot
   ```

For advanced Docker usage check out the extended documentation in:
- `utils/docker_build.md` - Comprehensive Docker deployment guide
- `utils/docker_build.sh` - Helper script for common Docker operations

Example using the helper script:
```bash
chmod +x utils/docker_build.sh  # Make script executable
./utils/docker_build.sh help    # Show available commands
```

## 📝 Commands

- `!calladitohelp` - Show available commands

*Basic interactions:*
- `!calladito @user` - Vote to mute the mentioned user in voice channels
- `!calladitoinfo` - Show all users with active votes
- `!calladitoinfo @user` - Show votes for a specific user
- `!calladitostatus` - Show mute system configuration

*Extra commands*
- `!calladitoclean @user` - (Admin only) Clear all votes against a user and unmute them if necessary
- `!calladitoping` - Check if the bot is active
- `!calladitodebug` - Show detailed information about the bot
- `!calladitoservidores` - Show servers where the bot is present

## 📊 Logging System

The bot automatically logs all actions to CSV files in the `logs` directory:

- Files are created daily in format `YYYY-MM-DD.csv`
- Each log entry contains: timestamp, action type, initiator, target, vote count, and guild ID
- Action types include: VOTE, MUTE, UNMUTE, and CLEAN
- Logs can be used for moderation auditing and statistics

## ⚙️ Advanced Configuration

You can modify the following mute and voting rules in the `main.go` file:

- `VOTE_DURATION` - Duration of votes (default: 30 minutes)
- `MUTE_DURATION` - Duration of voice muting (default: 1 minute)
- `VOTES_NEEDED` - Number of votes needed to mute a user (default: 1)

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome. Please open an issue or a pull request to suggest changes or improvements. 