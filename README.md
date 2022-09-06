# go-kafka-counter

Count kafka events from multiple topics.

This small utility reads kafka messages from multiple topics and provides you a
running count of the number of seen events. If the message is a json document,
you can also break down the counts for a topic by a specific key indicated by a
JSON path.

When you quit the program pressing Ctrl-C, it also provides some statistics of
the events seen broken down by topics.

This is a Golang rewrite of <https://github.com/sandipb/kafka-counter> which I abandoned
because of my application packaging requirements.
