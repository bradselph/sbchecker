# COD Status Bot User Guide

**Table of Contents**

* Introduction
* Getting Started
* Commands
  - /addaccount 
  - /removeaccount
  - /accountlogs
  - /updateaccount
  - /accountage
  - /setpreference
* Notifications
* Support

## Introduction

COD Status Bot is a Discord bot
designed to help you monitor your Activision accounts for any shadowban or permanent ban.
The bot periodically checks the status of your accounts (with a frequency of approximately **every 12 hours**)
and notifies you if there's a change in the ban status.

## Getting Started

To start using the bot, you'll need to add it to your Discord server. Here's how:

1. Invite the bot to your server using the provided [Invite Link](https://discord.com/oauth2/authorize?client_id=1211857854324015124).
2. Once the bot joins your server, it will automatically register the necessary commands.

## Commands

The bot provides several commands for you to interact with:

### /addaccount

Use this command to add a new account to be monitored by the bot. This should be your first command when you want to start monitoring an account.

**Usage:**

```
/addaccount <title> <sso_cookie>
```

- `<title>`: A name to identify the account with the bot and in the logs. This is for your own use, and the bot will also use this to identify the account in the logs and notifications.
- `<sso_cookie>`: The SSO (Single Sign-On) cookie associated with your Activision account.

**Example:**

```
/addaccount MyAccount 1234567890abcdef
```

**Note:** To get the SSO cookie, follow these steps:

1. Log in to your Activision account on a web browser.
2. Open the browser developer tools (the process may vary depending on your browser).
3. Navigate to the cookies section and find the cookie named `ACT_SSO_COOKIE` associated with the Activision domain.

**Important:** The SSO cookie grants access to your Activision account information. Keep it confidential and do not share it with anyone.

### /removeaccount

Use this command to remove an account from being monitored by the bot.

**Usage:**

```
/removeaccount <account>
```

- `<account>`: The title of the account you want to remove.

**Example:**

```
/removeaccount MyAccount
```

**Note:** Once you remove an account, all data associated with that account that the bot has stored or collected will be deleted. Be sure you want to remove the account before using this command as it cannot be undone. You will have to re-add the account if you want to monitor it again.

### /accountlogs

This command displays the last five shadowban logs for a specific account.

**Usage:**

```
/accountlogs <account>
```

- `<account>`: The title of the account you want to view logs for.

**Example:**

```
/accountlogs MyAccount
```

**Note:**

* If there are no logs available for the account, it may mean that the account was just added and hasn't been checked yet. The bot can only show the last five logs for an account after it was added to the bot. It cannot show logs from before the account was added for monitoring. This data does not normally exist as the bot itself creates this data from monitoring the account after it was added.
* At times, the bot may have a false positive or false negative result for an account, and the logs may show this. So, if a rapid change from good to shadowban or shadowban to good is seen in the logs, it may be a false positive or false negative result. If you see this, you can check the account manually to verify the status of the account or ignore the rapid change in the logs and assume the latest is the current status instead. This is normal and no need to worry about it.

The bot uses an internal flag system to track the status of your accounts and trigger notifications.
These logs are based on the bot's checks
and might not reflect real-time status compared to Activision.

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

**Notes:**

* When you update the SSO cookie for an account, the bot will use the new cookie for all future checks.
* This will not affect the logs for the account as the logs are stored separately from the account data. The logs will still show the status of the account based on the old cookie until the bot checks the account again and updates the logs with the new status of the account.
* This command is probably the most important as its function actually does the most work in the bot as it is responsible for updating the stored cookie as well as updating the flags for the account when it comes to checking and sending out notifications and ensuring the account data doesn't get corrupt during all this as well.

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

**Note:** The account age is calculated based on the date the account was created according to the Activision API and the current date and time when checked by the bot. This is important to note as the account age can be used to determine if an account is a new account or an old account as shadowbans are more common on new accounts than older accounts.

### /setpreference

**This command is currently dissabled as im still working on it**
This command allows you to set the preference for the bot to send notifications to the channel where the account was added or to your direct messages inbox (DMs).

**Usage:**

```
/setpreference <preference>
```

- `<preference>`: The preference for notifications. Valid values are `channel` or `dm`.

**Example:**

```
/setpreference channel
```

- If this preference is set, the bot will send notifications to the channel where the account was added. This is the default preference if no preference is set.

```
/setpreference dm
```

- If this preference is set, the bot will send notifications to your direct messages inbox.

**Notes:**

* The default preference is `channel`. If you have not set a preference, the bot will default to sending notifications to the channel where the account was added.
* You can only set the preference for your own account. The preference is stored per user, so it will apply to all accounts you add.
* You can change your preference at any time by using this command again.
* Setting the preference to `dm` will only send notifications to your direct messages inbox and not to the channel where the account was added. This is useful if you want to keep the notifications private and not share them with the rest of the server. If you want to share the notifications with the rest of the server, you can set the preference to 'channel' and the bot will send the notifications to the channel where the account was added.
### Important
  * Please note that you must use a channel to send commands as the bot does not respond to any messages or commands in the DMs. Only the notifications will go to your DMs if you set the preference to `dm`.

## Notifications

The bot will automatically send notifications:

* To the channel where the account was added (or to your DMs if you set the preference) whenever there's a change in the ban status of that account.
* Daily for each account, confirming that it's still being monitored.

You may also receive notifications regarding the validity of the SSO cookie for the account if the bot detects that the cookie is invalid or expired. This is to ensure that the bot can continue to monitor the account and send notifications for the account. If you receive this notification, you should update the SSO cookie for the account as soon as possible by using the /updateaccount command to ensure the bot can continue to monitor the account and send notifications for the account. If you do not wish to update the cookie for the account, you can remove the account from the bot by using the /removeaccount command. Otherwise, you may continue to receive notifications regarding the invalid or expired cookie for the account.

## Support

If you encounter any issues or have any questions, please don't hesitate to contact me or ask questions. I'm usually available on Discord as well as other webpages where you may have discovered this bot. I will be happy to help you with any issues or questions you may have regarding the bot or anything else you may need help with. I will do my best to help you with any issues or questions you may have and will try to respond as soon as possible. Thank you for using the bot, and I hope you find it useful and helpful for monitoring your accounts.
