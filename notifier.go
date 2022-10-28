package main

type apiBot interface {
	Send(chatID int64, text string)
}

type chatLister interface {
	ChatList() []int64
}

type Notifier struct {
	b apiBot
	l chatLister
}

func NewNotifier(b apiBot, l chatLister) Notifier {
	return Notifier{b: b, l: l}
}

func (n Notifier) NotifyAll(text string) {
	chats := n.l.ChatList()
	for _, c := range chats {
		n.b.Send(c, text)
	}
}
