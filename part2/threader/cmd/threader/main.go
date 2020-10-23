package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Pippolo84/go-services-patterns/part2/threader/internal/model"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

func main() {
	threadCmd := &cobra.Command{
		Use:   "thread [thread-id] [topic]",
		Short: "Create a new thread",
		Long:  `Create a new thread with the specified thread id and topic.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			tid, topic := args[0], args[1]
			if err := thread(tid, topic); err != nil {
				fmt.Fprintf(os.Stderr, "thread error: %v\n", err)
			}
		},
	}

	postCmd := &cobra.Command{
		Use:   "post [thread-id] [text]",
		Short: "Post a message to a thread",
		Long:  `Post a message to a thread identified by its thread id.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			tid, text := args[0], args[1]
			if err := post(tid, text); err != nil {
				fmt.Fprintf(os.Stderr, "post error: %v\n", err)
			}
		},
	}

	readCmd := &cobra.Command{
		Use:   "read [thread-id]",
		Short: "Read a thread",
		Long: `
Read all the messages of a thread identified by its thread id.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := read(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "read error: %v\n", err)
			}
		},
	}

	readAllCmd := &cobra.Command{
		Use:   "readall",
		Short: "Read all threads",
		Long: `
Read all the messages from all threads.`,
		Args: cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := readAll(); err != nil {
				fmt.Fprintf(os.Stderr, "readAll error: %v\n", err)
			}
		},
	}

	upvoteCmd := &cobra.Command{
		Use:   "upvote [thread-id] [message-id]",
		Short: "Upvote a message",
		Long: `
Upvote a message identified by its message id and the thread id of the thread it belongs to.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := upvote(args[0], args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "upvote error: %v\n", err)
			}
		},
	}

	rootCmd := &cobra.Command{
		Use:   "threader",
		Short: "threader is a thread manager to organize messages",
		Long: `
threader is a sample microservices application, used to illustrates some concepts
for the GoLab 2002 workshop "Design Patterns for production grade Go services".
See the GitHub repo for more info`,
	}
	rootCmd.AddCommand(threadCmd)
	rootCmd.AddCommand(postCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(readAllCmd)
	rootCmd.AddCommand(upvoteCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "args error: %v\n", err)
	}
}

func thread(tid, topic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(model.Thread{
		Topic: topic,
	}); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("http://localhost:8080/thread/%s", tid),
		bytes.NewReader(buf.Bytes()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("thread id: %s\n", tid)
		return nil
	}

	respBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(respBuf))

	return nil
}

func post(tid, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(model.Message{
		Text: text,
	}); err != nil {
		return err
	}

	mid := ksuid.New().String()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("http://localhost:8080/message/%s/%s", tid, mid),
		bytes.NewReader(buf.Bytes()),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("message id: %s\n", mid)
		return nil
	}

	respBuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(respBuf))

	return nil
}

func read(tid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://localhost:8080/thread/%s", tid),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var t model.Thread
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return err
	}

	fmt.Println(t.Topic)
	for _, msg := range t.Messages {
		fmt.Printf("\t%s\t+%d\t%s\n", msg.Timestamp, msg.Votes, msg.Text)
	}

	return nil
}

func readAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://localhost:8080/thread",
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var threads []model.Thread
	if err := json.NewDecoder(resp.Body).Decode(&threads); err != nil {
		return err
	}

	for _, t := range threads {
		fmt.Println(t.Topic)
		for _, msg := range t.Messages {
			fmt.Printf("\t%s\t+%d\t%s\n", msg.Timestamp, msg.Votes, msg.Text)
		}
	}

	return nil
}

func upvote(tid, mid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		fmt.Sprintf("http://localhost:8080/upvote/%s/%s", tid, mid),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("done!")
		return nil
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(buf))

	return nil
}
