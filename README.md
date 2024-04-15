### Prerequisites
- A Discord account and a server where you have permissions to add bots.
- Go programming language installed on your system.
- Basic knowledge of using the command line.

### Setup Instructions

1. **Clone the Repository**
   Clone the SBChecker repository from GitHub to your local machine using the following command:
   ```shell
   git clone https://github.com/bradselph/sbchecker.git
   ```

2. **Configure Environment Variables**
   Navigate to the cloned repository's directory and create a `.env` file with the following content:
   ```plaintext
   DISCORD_TOKEN=your_discord_bot_token
   DATABASE_URL=your_database_connection_string
   ```
   Replace `your_discord_bot_token` with the token you get from the Discord Developer Portal and `your_database_connection_string` with your database connection details.

3. **Install Dependencies**
   Install the necessary Go dependencies by running:
   ```shell
   go mod tidy
   ```

4. **Build the Bot**
   Compile the bot using the Go compiler:
   ```shell
   go build -o sbchecker
   ```

5. **Run the Bot**
   Start the bot by executing the compiled binary:
   ```shell
   ./sbchecker
   ```

### Usage Instructions

1. **Adding the Bot to Your Server**
   - Go to the Discord Developer Portal and create a new application.
   - Under the "Bot" tab, click "Add Bot".
   - Use the generated token in your `.env` file.
   - Under the "OAuth2" tab, select the appropriate scopes and permissions for your bot.
   - Use the generated OAuth2 URL to invite the bot to your server.

2. **Registering Commands**
   - The bot will automatically register slash commands upon joining a server.
   - You can use the `/addaccount`, `/removeaccount`, `/accountlogs`, and `/updateaccount` commands to manage accounts.

3. **Monitoring Accounts**
   - Use the `/addaccount` command to register an account for monitoring.
   - The bot will periodically check the status of the account and notify you if there are any changes.

4. **Receiving Notifications**
   - When an account status changes, the bot will send a message to the designated Discord channel with details about the change.

5. **Viewing Account Logs**
   - Use the `/accountlogs` command to view the history of status changes for an account.

6. **Updating and Removing Accounts**
   - Use the `/updateaccount` command to update account details.
   - Use the `/removeaccount` command to stop monitoring an account.

### Additional Information
- Ensure that your bot has the necessary permissions to send messages and manage channels in your Discord server.
- For detailed command usage and additional configuration options, refer to the `README.md` file in the GitHub repository.

By following these instructions, you should be able to set up and use the SBChecker Discord bot to monitor account statuses effectively. If you encounter any issues, check the bot's logs for error messages that can help you troubleshoot the problem.
