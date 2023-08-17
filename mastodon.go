package main

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Account struct {
	ID                  string   `json:"id"`
	Acct                string   `json:"acct"`
	Avatar              string   `json:"avatar"`
	AvatarStatic        string   `json:"avatarStatic"`
	Bot                 bool     `json:"bot"`
	CreatedAt           string   `json:"createdAt"`
	Discoverable        bool     `json:"discoverable"`
	DisplayName         string   `json:"displayName"`
	Emojis              []Emoji  `json:"emojis"`
	Fields              []any    `json:"fields"`
	FollowRequestsCount int      `json:"followRequestsCount"`
	FollowersCount      int      `json:"followersCount"`
	FollowingCount      int      `json:"followingCount"`
	FQN                 string   `json:"fqn"`
	Header              string   `json:"header"`
	HeaderStatic        string   `json:"headerStatic"`
	LastStatusAt        *string  `json:"lastStatusAt"`
	Locked              bool     `json:"locked"`
	Note                string   `json:"note"`
	Roles               []string `json:"roles"`
	Source              *Source  `json:"source"`
	StatusesCount       int      `json:"statusesCount"`
	URL                 string   `json:"url"`
	Username            string   `json:"username"`
}

type Status struct {
	ID                 string       `json:"id"`
	Account            *Account     `json:"account"`
	Card               *PreviewCard `json:"card"`
	Content            string       `json:"content"`
	CreatedAt          string       `json:"createdAt"`
	InReplyToID        *string      `json:"inReplyToId"`
	InReplyToAccountID *string      `json:"inReplyToAccountId"`
	Sensitive          bool         `json:"sensitive"`
	SpoilerText        string       `json:"spoilerText"`
	Visibility         string       `json:"visibility"`
	Language           string       `json:"language"`
	RepliesCount       int          `json:"repliesCount"`
	ReblogsCount       int          `json:"reblogsCount"`
	FavouritesCount    int          `json:"favouritesCount"`
	Favourited         bool         `json:"favourited"`
	Reblogged          bool         `json:"reblogged"`
	Muted              bool         `json:"muted"`
	Bookmarked         bool         `json:"bookmarked"`
	Reblog             any          `json:"reblog"`
	Application        any          `json:"application"`
	MediaAttachments   []Attachment `json:"mediaAttachments"`
	Mentions           []Mention    `json:"mentions"`
	Tags               []any        `json:"tags"`
	Emojis             []Emoji      `json:"emojis"`
	Poll               any          `json:"poll"`
	URI                string       `json:"uri"`
	URL                string       `json:"url"`
}

type Source struct {
	Fields              []any  `json:"fields"`
	Language            string `json:"language"`
	Note                string `json:"note"`
	Privacy             string `json:"privacy"`
	Sensitive           bool   `json:"sensitive"`
	FollowRequestsCount int    `json:"followRequestsCount"`
}

type Mention struct {
	ID       string `json:"id"`
	Acct     string `json:"acct"`
	Username string `json:"username"`
	URL      string `json:"url"`
}

type Emoji struct {
	Shortcode string `json:"shortcode"`
	StaticURL string `json:"staticURL"`
	URL       string `json:"url"`
}

type PreviewCard struct {
	URL          string  `json:"url"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Type         string  `json:"type"`
	AuthorName   string  `json:"author_name"`
	AuthorURL    string  `json:"author_url"`
	ProviderName string  `json:"provider_name"`
	ProviderURL  string  `json:"provider_url"`
	HTML         string  `json:"html"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Image        string  `json:"image"`
	EmbedURL     string  `json:"embed_url"`
	Blurhash     *string `json:"blurhash"`
}

type Attachment struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	URL         string         `json:"uRL"`
	PreviewURL  string         `json:"preview_url"`
	RemoteURL   string         `json:"remote_url"`
	Meta        map[string]any `json:"meta"`
	Description string         `json:"description"`
	Blurhash    string         `json:"blurhash"`
}

type ToAccountOpts struct {
	WithSource bool
}

func toAccount(ctx context.Context, p *Profile, opts *ToAccountOpts) *Account {
	createdAt := ""
	if p.event == nil {
		createdAt = p.event.CreatedAt.Time().Format(time.RFC3339)
	}

	account := Account{
		ID:                  p.pubkey,
		Acct:                p.NIP05,
		Avatar:              p.Picture,
		AvatarStatic:        p.Picture,
		Bot:                 false,
		CreatedAt:           createdAt,
		Discoverable:        true,
		DisplayName:         p.Name,
		Emojis:              toEmojis(p.event),
		Fields:              []any{},
		FollowRequestsCount: 0,
		FollowersCount:      0,
		FollowingCount:      0,
		FQN:                 p.handle(),
		Header:              p.Banner,
		HeaderStatic:        p.Banner,
		LastStatusAt:        nil,
		Locked:              false,
		Note:                html.EscapeString(p.About),
		Roles:               []string{},
		Source:              nil,
		StatusesCount:       0,
		URL:                 "http://" + srv.Addr + "/users/" + p.pubkey,
		Username:            p.handle(),
	}

	if opts != nil && opts.WithSource {
		account.Source = &Source{
			Fields:              []any{},
			Language:            "",
			Note:                p.About,
			Privacy:             "public",
			Sensitive:           false,
			FollowRequestsCount: 0,
		}
	}

	return &account
}

func toMention(ctx context.Context, pubkey string) Mention {
	profile := loadProfile(ctx, pubkey)
	if profile != nil {
		if account := toAccount(ctx, profile, nil); account != nil {
			return Mention{
				ID:       account.ID,
				Acct:     account.Acct,
				Username: account.Username,
				URL:      account.URL,
			}
		}
	}

	npub, _ := nip19.EncodePublicKey(pubkey)
	return Mention{
		ID:       pubkey,
		Acct:     npub,
		Username: npub[:8],
		URL:      "http://" + srv.Addr + "/users/" + pubkey,
	}
}

func toStatus(ctx context.Context, evt *nostr.Event) *Status {
	profile := loadProfile(ctx, evt.PubKey)

	var account *Account
	if profile != nil {
		account = toAccount(ctx, profile, nil)
	}

	replyTag := evt.Tags.GetFirst([]string{"e", ""})
	var inReplyToId *string
	if replyTag != nil {
		inReplyToId = &(*replyTag)[1]
	}

	mentionedPubkeys := make(map[string]bool, len(evt.Tags))
	mentions := make([]Mention, 0, len(evt.Tags))
	for _, tag := range evt.Tags {
		if tag[0] == "p" {
			pubkey := tag[1]
			if _, exists := mentionedPubkeys[pubkey]; !exists {
				mentionedPubkeys[pubkey] = true
				mentions = append(mentions, toMention(ctx, pubkey))
			}
		}
	}

	attachments := make([]Attachment, 0, 5)
	for _, link := range urlMatcher.FindAllString(evt.Content, -1) {
		u, err := url.Parse(link)
		if err != nil {
			continue
		}

		attachmentType := ""
		switch {
		case strings.HasSuffix(u.Path, ".mp4"):
			attachmentType = "video"
		case strings.HasSuffix(u.Path, ".webm"):
			attachmentType = "video"
		case strings.HasSuffix(u.Path, ".gifv"):
			attachmentType = "gifv"
		case strings.HasSuffix(u.Path, ".mp3"):
			attachmentType = "audio"
		case strings.HasSuffix(u.Path, ".ogg"):
			attachmentType = "audio"
		case strings.HasSuffix(u.Path, ".webp"):
			attachmentType = "image"
		case strings.HasSuffix(u.Path, ".jpg"):
			attachmentType = "image"
		case strings.HasSuffix(u.Path, ".jpeg"):
			attachmentType = "image"
		case strings.HasSuffix(u.Path, ".gif"):
			attachmentType = "image"
		case strings.HasSuffix(u.Path, ".png"):
			attachmentType = "image"
		default:
			continue
		}

		attachments = append(attachments, Attachment{
			ID:          link,
			Type:        attachmentType,
			URL:         link,
			PreviewURL:  link,
			RemoteURL:   "",
			Meta:        map[string]interface{}{},
			Description: "",
			Blurhash:    "",
		})
	}

	text := evt.Content
	if len(mentions) > 0 {
		elements := make([]string, len(mentions))
		for i, mention := range mentions {
			username := mention.Username
			if strings.HasPrefix(username, "npub1") {
				username = username[:8]
			}
			elements[i] = fmt.Sprintf(`<a href="%s" class="u-url mention" rel="ugc">@<span>%s</span></a>`, mention.URL, username)
		}
		text = fmt.Sprintf(`<span class="recipients-inline">%s</span>`, strings.Join(elements, " "))
	}

	cw := evt.Tags.GetFirst([]string{"content-warning", ""})
	cwText := ""
	if cw != nil {
		cwText = (*cw)[1]
	}

	return &Status{
		ID:                 evt.ID,
		Account:            account,
		Card:               nil,
		Content:            text,
		CreatedAt:          evt.CreatedAt.Time().Format(time.RFC3339),
		InReplyToID:        inReplyToId,
		InReplyToAccountID: nil,
		Sensitive:          cw != nil,
		SpoilerText:        cwText,
		Visibility:         "public",
		Language:           "",
		RepliesCount:       0,
		ReblogsCount:       0,
		FavouritesCount:    0,
		Favourited:         false,
		Reblogged:          false,
		Muted:              false,
		Bookmarked:         false,
		Reblog:             nil,
		Application:        nil,
		MediaAttachments:   attachments,
		Mentions:           mentions,
		Tags:               nil,
		Emojis:             toEmojis(evt),
		Poll:               nil,
		URI:                "http://" + srv.Addr + "/posts/" + evt.ID,
		URL:                "http://" + srv.Addr + "/posts/" + evt.ID,
	}
}

// func unfurlCard(url string) (*PreviewCard, error) {
// 	fmt.Printf("Unfurling %s...\n", url)
// 	result, err := unfurl(url, unfurl.Options{
// 		Fetch:   fetch,
// 		Follow:  2,
// 		Timeout: Time.Seconds(1),
// 		Size:    1024 * 1024,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	card := &PreviewCard{
// 		Type:         result.OEmbed.Type,
// 		URL:          result.CanonicalURL,
// 		Title:        result.OEmbed.Title,
// 		Description:  result.OpenGraph.Description,
// 		AuthorName:   result.OEmbed.AuthorName,
// 		AuthorURL:    result.OEmbed.AuthorURL,
// 		ProviderName: result.OEmbed.ProviderName,
// 		ProviderURL:  result.OEmbed.ProviderURL,
// 		HTML:         sanitizeHTML(result.OEmbed.HTML),
// 		Width:        result.OEmbed.Width,
// 		Height:       result.OEmbed.Height,
// 		Image:        result.OEmbed.Thumbnails[0].URL,
// 		EmbedURL:     "",
// 		Blurhash:     nil,
// 	}
//
// 	return card, nil
// }
//
// func unfurlCardCached(url string) (*PreviewCard, error) {
// 	cached, ok := previewCache.Get(url)
// 	if ok {
// 		return cached.(*PreviewCard), nil
// 	}
//
// 	card, err := unfurlCard(url)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	previewCache.Set(url, card)
//
// 	return card, nil
// }

func toEmojis(event *nostr.Event) []Emoji {
	emojiTags := make([][]string, 0, len(event.Tags))
	for _, tag := range event.Tags {
		if tag[0] == "emoji" {
			emojiTags = append(emojiTags, tag)
		}
	}

	emojis := make([]Emoji, len(emojiTags))
	for i, tag := range emojiTags {
		emojis[i] = Emoji{
			Shortcode: tag[1],
			StaticURL: tag[2],
			URL:       tag[2],
		}
	}

	return emojis
}
