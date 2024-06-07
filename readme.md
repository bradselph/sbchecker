# COD Status Bot User Guide

## Introduction

COD Status Bot is a Discord bot designed to help you monitor your Activision accounts for any shadowban or permanent ban. The bot periodically checks the status of your accounts and notifies you if there's a change in the ban status.

## Getting Started

Before you can start using the bot, you'll need to add it to your Discord server. Follow these steps:

1. Invite the bot to your server using the provided invite link.
2. Once the bot joins your server, it will automatically register the necessary commands.

## Commands

The bot provides the following commands for you to interact with:

### /addaccount

This command allows you to add a new account to be monitored by the bot.

**Usage:**

```
/addaccount <title> <sso_cookie>
```

- `<title>`: A descriptive title for your account.
- `<sso_cookie>`: The SSO (Single Sign-On) cookie associated with your Activision account.

**Example:**

```
/addaccount MyAccount 1234567890abcdef
```

### /removeaccount

This command allows you to remove an account from being monitored by the bot.

**Usage:**

```
/removeaccount <account>
```

- `<account>`: The title of the account you want to remove.

**Example:**

```
/removeaccount MyAccount
```

### /accountlogs

This command displays the last 5 shadowban logs for a specific account.

**Usage:**

```
/accountlogs <account>
```

- `<account>`: The title of the account you want to view logs for.

**Example:**

```
/accountlogs MyAccount
```

### /updateaccount

This command allows you to update the SSO cookie for an existing account.

**Usage:**

```
/updateaccount <account> <sso_cookie>
```

- `<account>`: The title of the account you want to update.
- `<sso_cookie>`: The new SSO cookie for the account.

**Example:**

```
/updateaccount MyAccount abcdef1234567890
```

### /accountage

This command displays the age of a specific account.

**Usage:**

```
/accountage <account>
```

- `<account>`: The title of the account you want to check the age for.

**Example:**

```
/accountage MyAccount
```

## Notifications

The bot will automatically send notifications to the channel where the account was added whenever there's a change in the ban status of that account. Additionally, the bot will send a daily update for each account, confirming that it's still being monitored.

## Support

If you encounter any issues or have any questions, please reach out to the bot's support team for assistance.
