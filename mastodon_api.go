package main

import (
	"context"
	"html"
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
	Id                 string       `json:"id"`
	Account            *Account     `json:"account"`
	Card               *PreviewCard `json:"card"`
	Content            string       `json:"content"`
	CreatedAt          string       `json:"createdAt"`
	InReplyToId        *string      `json:"inReplyToId"`
	InReplyToAccountId *string      `json:"inReplyToAccountId"`
	Sensitive          string       `json:"sensitive"`
	SpoilerText        string       `json:"spoilerText"`
	Visibility         string       `json:"visibility"`
	Language           string       `json:"language"`
	RepliesCount       string       `json:"repliesCount"`
	ReblogsCount       int          `json:"reblogsCount"`
	FavouritesCount    int          `json:"favouritesCount"`
	Favourited         bool         `json:"favourited"`
	Reblogged          bool         `json:"reblogged"`
	Muted              bool         `json:"muted"`
	Bookmarked         bool         `json:"bookmarked"`
	Reblog             any          `json:"reblog"`
	Application        any          `json:"application"`
	MediaAttachments   string       `json:"mediaAttachments"`
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
	Id       string `json:"id"`
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

type ToAccountOpts struct {
	WithSource bool
}

func (p Profile) toAccount(ctx context.Context, opts *ToAccountOpts) (*Account, error) {
	account := Account{
		ID:                  p.pubkey,
		Acct:                p.NIP05,
		Avatar:              p.Picture,
		AvatarStatic:        p.Picture,
		Bot:                 false,
		CreatedAt:           p.event.CreatedAt.Time().Format(time.RFC3339),
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

	if opts.WithSource {
		account.Source = &Source{
			Fields:              []any{},
			Language:            "",
			Note:                p.About,
			Privacy:             "public",
			Sensitive:           false,
			FollowRequestsCount: 0,
		}
	}

	return &account, nil
}

func toMention(ctx context.Context, pubkey string) Mention {
	profile := loadProfile(ctx, pubkey)
	if profile != nil {
		if account, err := profile.toAccount(ctx, nil); err == nil {
			return Mention{
				Id:       account.ID,
				Acct:     account.Acct,
				Username: account.Username,
				URL:      account.URL,
			}
		}
	}

	npub, _ := nip19.EncodePublicKey(pubkey)
	return Mention{
		Id:       pubkey,
		Acct:     npub,
		Username: npub[:8],
		URL:      "http://" + srv.Addr + "/users/" + pubkey,
	}
}

//	func toStatus(ctx context.Context, event *nostr.Event) (*Status, error) {
//		profile, err := loadProfile(ctx, event.PubKey)
//		if err != nil {
//			return nil, err
//		}
//
//		var account *Account
//		if profile != nil {
//			account, err = profile.toAccount(ctx, nil)
//			if err != nil {
//				return nil, err
//			}
//		}
//
//		replyTag := event.Tags.GetFirst([]string{"e", ""})
//		var inReplyToId *string
//		if replyTag != nil {
//			inReplyToId = &(*replyTag)[1]
//		}
//
//		mentionedPubkeys := make(map[string]bool, len(event.Tags))
//		for _, tag := range event.Tags {
//			if tag[0] == "p" {
//				mentionedPubkeys[tag[1]] = true
//			}
//		}
//
//		html, links, firstURL := parseNoteContent(event.Content)
//		mediaLinks := getMediaLinks(links)
//
//		mentions := make([]Mention, len(mentionedPubkeys))
//		for pubkey := range mentionedPubkeys {
//			mentions[i] = toMention(ctx, pubkey)
//		}
//
//		var card *PreviewCard
//		if firstURL != "" {
//			card, err = unfurlCardCached(ctx.firstURL)
//			if err != nil {
//				return nil, err
//			}
//		}
//
//		content := buildInlineRecipients(mentions) + html
//
//		cw := findCWTag(event)
//		subject := findSubjectTag(event)
//
//		return &Status{
//			Id:                 event.ID,
//			Account:            account,
//			Card:               card,
//			Content:            content,
//			CreatedAt:          event.CreatedAt.Time().Format(time.RFC3339),
//			InReplyToId:        inReplyToId,
//			InReplyToAccountId: nil,
//			Sensitive:          cw != nil,
//			SpoilerText:        cw[1],
//			Visibility:         "public",
//			Language:           "",
//			RepliesCount:       0,
//			ReblogsCount:       0,
//			FavouritesCount:    0,
//			Favourited:         false,
//			Reblogged:          false,
//			Muted:              false,
//			Bookmarked:         false,
//			Reblog:             nil,
//			Application:        nil,
//			MediaAttachments:   renderAttachments(mediaLinks),
//			Mentions:           mentions,
//			Tags:               nil,
//			Emojis:             toEmojis(event),
//			Poll:               nil,
//			URI:                "http://" + srv.Addr + "/posts/" + event.ID,
//			URL:                "http://" + srv.Addr + "/posts/" + event.ID,
//		}, nil
//	}
//
//	func buildInlineRecipients(mentions []map[string]string) string {
//		if len(mentions) == 0 {
//			return ""
//		}
//
//		elements := make([]string, len(mentions))
//		for i, mention := range mentions {
//			username := mention["username"]
//			if nip19.BECH32_REGEX.MatchString(username) {
//				username = username[:8]
//			}
//			elements[i] = fmt.Sprintf(`<a href="%s" class="u-url mention" rel="ugc">@<span>%s</span></a>`, mention["url"], username)
//		}
//
//		return fmt.Sprintf(`<span class="recipients-inline">%s </span>`, strings.Join(elements, " "))
//	}
//
//	func renderAttachments(mediaLinks []MediaLink) []map[string]interface{} {
//		attachments := make([]map[string]interface{}, len(mediaLinks))
//		for i, link := range mediaLinks {
//			baseType := strings.Split(link.MimeType, "/")[0]
//			attachmentType := attachmentTypeSchema.Parse(baseType).(string)
//
//			attachments[i] = map[string]interface{}{
//				"id":          link.URL,
//				"type":        attachmentType,
//				"url":         link.URL,
//				"preview_url": link.URL,
//				"remote_url":  nil,
//				"meta":        map[string]interface{}{},
//				"description": "",
//				"blurhash":    nil,
//			}
//		}
//
//		return attachments
//	}
//
//	func unfurlCard(url string) (*PreviewCard, error) {
//		fmt.Printf("Unfurling %s...\n", url)
//		result, err := unfurl(url, unfurl.Options{
//			Fetch:   fetch,
//			Follow:  2,
//			Timeout: Time.Seconds(1),
//			Size:    1024 * 1024,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		card := &PreviewCard{
//			Type:         result.OEmbed.Type,
//			URL:          result.CanonicalURL,
//			Title:        result.OEmbed.Title,
//			Description:  result.OpenGraph.Description,
//			AuthorName:   result.OEmbed.AuthorName,
//			AuthorURL:    result.OEmbed.AuthorURL,
//			ProviderName: result.OEmbed.ProviderName,
//			ProviderURL:  result.OEmbed.ProviderURL,
//			HTML:         sanitizeHTML(result.OEmbed.HTML),
//			Width:        result.OEmbed.Width,
//			Height:       result.OEmbed.Height,
//			Image:        result.OEmbed.Thumbnails[0].URL,
//			EmbedURL:     "",
//			Blurhash:     nil,
//		}
//
//		return card, nil
//	}
//
//	func unfurlCardCached(url string) (*PreviewCard, error) {
//		cached, ok := previewCache.Get(url)
//		if ok {
//			return cached.(*PreviewCard), nil
//		}
//
//		card, err := unfurlCard(url)
//		if err != nil {
//			return nil, err
//		}
//
//		previewCache.Set(url, card)
//
//		return card, nil
//	}

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
