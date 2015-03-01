# twitter-archive

Simple CLI tool for archiving a user's entire Twitter feed.

`twitter-archive` has two basic modes:

 - *archive mode*—archive as much (usually the last 3,200 tweets) of a
   user's timeline;
 - *update mode*—given a user's existing archive prepend upto 3,200 of
   the user's most recent tweets.

 `twitter-archive` fetches the entire JSON object associated with each tweet in a user's timeline, upto the maximum number of tweets Twitter will allow to be fetched (currently 3,200).

 Take a look [here](https://dev.twitter.com/rest/reference/get/statuses/show/%3Aid) to see what the schema of a tweet looks like.

`twitter-archive` can optionally exclude *retweets* and *@-mentions* from the archived tweets.

## Setting Up

You will need to create a new application for Twitter at [https://apps.twitter.com/](https://apps.twitter.com/).
Make a note of the `API Key` and `API secret`.
You will also need to add yourself (your Twitter user) to the Application, which will then provide you with an `Access Token` and an `Access Token Secret`.

`twitter-archive` will need these tokens either as environment variables or via a `JSON` configuration file.
Environment variables are mapped as follows:

 - `TWITTER_CONSUMER_KEY` Application API Key;
 - `TWITTER_CONSUMER_SECRET` Application API Secret;
 - `TWITTER_ACCESS_TOKEN` Access Token;
 - `TWITTER_ACCESS_TOKEN_SECRET` Access Token Secret;

 Alternatively you can provide a JSON file that looks like:

 ```json
{
  "consumer_key": "S",
  "consumer_secret": "",
  "access_token": "",
  "access_token_secret": ""
}
 ```

By default `twitter-archive` will look for a `.taconfig` file containing this JSON object, from the directory you are running the program.
You can specify the location of this file with the `-c` option.

## Install

Assuming you have the Go compiler on your machine:

```
go get github.com/e-dard/twitter-archive
```

`twitter-archive` will then be available in `$GOPATH/bin`, which you already have in your `$PATH`, right?

## Usage
`twitter-archive` outputs to `stdout`, so you're most likely going to want to redirect the output to a file.

### Archiving Tweets

Getting a fresh archive of a user's timeline:

```
$ twitter-archive ThisisPartridge
{...}
{...}
{...}
```

If you don't want to include *@-mentions* or *retweets* in the archive, i.e., you just want a user's first-person status updates:

```
$ twitter-archive -noat -nort ThisisPartridge
{...}
{...}
{...}
```

### Keeping an archive updated
Twitter only lets you get at the last 3,200 of a user's tweet.
Therefore to keep an archive of more than that you will need to regularly request the latest tweets for a user.
The *update* mode allows you to do that:

```
$ twitter-archive -a /path/to/existing/archivefile
```

where `archivefile` would be a file containing previously archived tweets.
`twitter-archive` will read the most recent tweet in the archive to figure out what point it needs to fetch tweets upto.
It will also use the archive to determine the user that the timeline belongs to, hence why you don't need to provide the `user` argument.
You will still need to manually provide the `-noat` and `-nort` options.

## Combining twitter-archive with other things

Let's say you only want the text from a user's tweets:

```
twitter-archive ThisisPartridge | json -ga -e "this.text = this.text.replace('\n', ' ')" text
```

This example relied on you having the [json](http://trentm.com/json/) tool installed.
Note that it replaces any line-breaks in tweets with a space, so that you guarantee to end up with each tweet on its own line.
