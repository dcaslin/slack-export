# Slack Export

- [Slack Export](#slack-export)
  - [Summary](#summary)
  - [Use case](#use-case)
  - [Parameters](#parameters)
  - [Running the application](#running-the-application)
    - [Command line](#command-line)
    - [Docker](#docker)
  - [Creating your App and Obtaining a Token](#creating-your-app-and-obtaining-a-token)
  - [Developing](#developing)



## Summary

This is a utility, based on [rohanshekhar's fork of Slack Export](https://github.com/rohanshekhar/slack-export). It exports all accessible slack history in to text files (plus, optionally JSON files and the messages attached files).

**This tool requires an API TOKEN, see below for detailed instructions, valid as of April 2021, for obtaining one.** You must be a workspace admin to use this.


## Use case

This tool was created to allow me to search archives on a small personal Slack workspace, a workspace that replaced old Google group chats that was I was used to being able to search years in the past. In April 2021 there were precious few gists or Github repos that did not rely on defunct Slack API functionality, so I built this one. My use case steps

- I upgraded to a paid plan temporarily
- Ran this tool to export all my history
- Then setup a schedule task to run backups on a weekly basis. The tool keeps its own state of the last message export date for each channel/DM, so after your initial export future exports should be small and quick.
- The backups are targeted to a folder that's mapped to Google Drive, so Google Drive searches will also query the text/file archives of my Slack chats 

Notes: 

- This was not intended nor tested with a large commercial Slack workspace, if you need to export something like that I'd suggest actually upgrading to Slack Export it and staying in the Sandbox.

- This technique id blessed by slack. https://get.slack.help/hc/en-us/articles/204897248
  
## Parameters

All parameters for this tool are set via Environmental variables. See `sample.env` for a full list of example params: 

- `SLACK_API_TOKEN` A token that starts with "xox" that you can read how to obtain below. |Required|
- `EXPORT_ROOT` The target folder where all exports will be written |Required|
- `EXPORT_JSON` If true, JSON payloads will be exported next to their respective file (`mychannel-date.txt` and `mychannel-date.json`) |Optional, default false|
- `EXPORT_FILES` If true, files attached to messages will be downloaded into subfolders (`root/mychannel/1234-file.jpg`) This can significantly slow down initial exports |Optional, default false|
- `MIN_DATE_ISO` Typically only used for dev, this ISO formatted string hard sets the minimum date of a file to export rather than grabbing all accessible messages |Optional, default null|
- `MAX_DATE_ISO` Same as above, only for the maximum date to retrieve |Optional, default null|
- `RESET_CONF` If true all state is forgotten. This will cause a full export on next run. If you set this to true, remember to remove it later so you're not constantly downloading everything |Optional, default false|

## Running the application

### Command line
- Install node.js (this was developed and tested with Node 14 LTS, but it's probably not very sensitive)
- Clone this repo
- `npm install`
- Set your environmental variables or a proper .env file at the root (see `sample.env`)
- `npm start`

### Docker

This is hosted on docker hub at https://hub.docker.com/repository/docker/dcaslin/slack-export

- Update the reference `docker-compose.yaml` file to, at a minimum, set your `SLACK_API_TOKEN` and make sure the export bind mount exists with proper permissions
- Run `docker-compose up`. This will run then CLI tool in a container that will immediately shut down on success

## Creating your App and Obtaining a Token

To use this application, you'll need a token. There are a lot of tools out there that require a "Legacy App Token" which you *cannot get anymore*. This tool uses a proper modern Slack App token. When you're done you should have a string that starts with 

1. Go to https://api.slack.com/apps and click the "Create New App" button
   ![image](https://user-images.githubusercontent.com/16437968/115113145-c7061300-9f56-11eb-8cfa-87222a7407e2.png)
2. Name it whatever you want. Choose your target workspace as your "Development Slack Workspace". Click "Create App"
   ![image](https://user-images.githubusercontent.com/16437968/115113521-d5edc500-9f58-11eb-977d-444fc4e481f5.png)
3. Click "Permissions" to start adding out permissions for our app
   ![image](https://user-images.githubusercontent.com/16437968/115112586-29114900-9f54-11eb-941e-1144d3285121.png)
4. Add out the OAuth user scopes we need
   ![image](https://user-images.githubusercontent.com/16437968/115112667-9329ee00-9f54-11eb-99fb-46d9b232734d.png)
5. We'll want the following 10 permissions (note, this is not thoroughly tested, you might be able to get by with less)
   - channels:history
   - channels:read
   - files:read
   - groups:history
   - groups:read
   - im:history
   - im:read
   - mpim:history
   - mpim:read
   - users:read
6. Now let's install our app
   ![image](https://user-images.githubusercontent.com/16437968/115112962-df296280-9f55-11eb-95e5-540d7ff4f14c.png)
   ![image](https://user-images.githubusercontent.com/16437968/115113270-81961580-9f57-11eb-8c18-a135ef63dd13.png)
7. And now we finally have our token! Copy this, we'll need it for our app
   ![image](https://user-images.githubusercontent.com/16437968/115113308-adb19680-9f57-11eb-8702-d2e0219db0f3.png)

## Developing

This is built in Node and Typescript.

- clone this repo
- Run `npm install`
- Create a .env file with, at minimum `SLACK_API_TOKEN=[your api token]`
- Run `npm run start:dev`