# Threader

Threader is a simple application to manage threads and messages in a microservices way.

- thread <thread_id_1>
  - message <message_id_1>
  - message <message_id_2>
  - message <message_id_3>
  - ...
- thread <thread_id_2>
  - message <message_id_1>
  - message <message_id_2>
  - message <message_id_3>
  - ...
...

It has three sub-components: archive, broker and threader.

### **archive**

It is used to gather analytics (just the score) of the messages in the system.

1. Modify the score of a message, identified by its `message_id` and `thread_id`

```
curl --request POST \
  --url http://localhost:8081/score \
  --header 'content-type: application/json' \
  --data '{
	"thread_id": "1",
	"message_id": "1",
	"votes": 10
}'
```

2. Get the message with the highest score

```
url --request GET \
  --url http://localhost:8081/highest-score

{"thread_id":"1","message_id":"1","votes":10}
```

### **broker**

Used to aggregate all threads and messages from the users.

1. List all threads

```
curl --request GET   --url http://localhost:8080/thread

[
  {
    "id": "1",
    "topic": "test-topic",
    "messages": [
      {
        "id": "1",
        "thread_id": "1",
        "timestamp": "2020-10-23T09:41:42+02:00",
        "text": "most smart text ever",
        "votes": 1
      },
      {
        "id": "1",
        "thread_id": "1",
        "timestamp": "2020-10-23T09:47:01+02:00",
        "text": "not so smart text",
        "votes": 0
      }
    ]
  }
]

```

2. Get a single thread identified by its thread id

```
curl --request GET \
  --url http://localhost:8080/thread/1

{
  "id": "1",
  "topic": "test-topic",
  "messages": [
    {
      "id": "1",
      "thread_id": "1",
      "timestamp": "2020-10-23T09:41:42+02:00",
      "text": "most smart text ever 1",
      "votes": 1
    },
    {
      "id": "1",
      "thread_id": "1",
      "timestamp": "2020-10-23T09:47:01+02:00",
      "text": "not so smart text",
      "votes": 0
    }
  ]
}
```

3. Post a new thread

```
curl --request POST \
  --url http://localhost:8080/thread/1 \
  --header 'content-type: application/json' \
  --data '{
	"topic": "test-topic"
}'
```

4. Post a new message

```
curl --request POST \
  --url http://localhost:8080/message/1/1 \
  --header 'content-type: application/json' \
  --data '{
	"text": "most smart text ever"
}'
```

5. Upvote a message, identified by a thread id and a message id

```
curl --request PATCH \
  --url http://localhost:8080/upvote/1/1
```

The broker communicates with the archive every time:

- it creates a new message (setting the score of the message to 0)
- it receives a request to upvote a message

### **threader**

A CLI application to interact with the broker manually.

```
$ bin/threader

threader is a sample microservices application, used to illustrates some concepts
for the GoLab 2002 workshop "Design Patterns for production grade Go services".
See the GitHub repo for more info

Usage:
  threader [command]

Available Commands:
  help        Help about any command
  post        Post a message to a thread
  read        Read a thread
  readall     Read all threads
  thread      Create a new thread
  upvote      Upvote a message

Flags:
  -h, --help   help for threader

Use "threader [command] --help" for more information about a command.
```

Example of a new thread creation:

```
$ bin/threader thread 1 "What do you think about this workshop?"
thread id: 1
```

Examples of some new messages posted:

```
$ bin/threader post 1 "I don't care, I just want to have lunch"
message id: 1jGoTV0F3BThB0wLT9p8s4j3OpQ

$ bin/threader post 1 "Meh"
message id: 1jGoW0XreeOE8FMxm4ZwEDRmV3A

$ bin/threader post 1 "From now, I am gonna write my services in Node.js"
message id: 1jGoXxNjNNp8aIUodkS4pLrBV2n
```

Please note how the `threader` application autogenerates the messages id.

Example of an upvote of a message:

```
$ bin/threader upvote 1 1jGoXxNjNNp8aIUodkS4pLrBV2n
done!
```

Example of listing all threads and messages:

```
$ bin/threader readall
What do you think about this workshop?
	2020-10-23T09:56:36+02:00	+0	I don't care, I just want to have lunch
	2020-10-23T09:56:56+02:00	+0	Meh
	2020-10-23T09:57:12+02:00	+1	From now, I am gonna write my services in Node.js
```

### Requisites:

Please note that the **archive** service is failing (pseudo) randomly.
The application should be more resilient. Specifically:

- it should try to recover from a transient failure
- it should be responsive and fail fast if the failure is not so transient
- it should avoid hammering the **archive** too much in case of a failure

### What you should do:

1. modify the **broker** code to avoid hammering the **archive** when it is failing (choose the strategy you prefer and experiment with it)
2. build the services and try them (directly or through the **threader** CLI)
3. verify that you don't wait indefinitely (or too much) for a response from the **broker** when the **archive** fails
4. verify that you are not hammering the **archive** and worsening its shape
5. Have fun!

### Time to spare?

Why not add some unit and integration tests to the services? :D