package mock

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/wirepair/gcd/v2/gcdapi"
	"gitlab.com/browserker/browserk"
)

func MakeMockConfig() *browserk.Config {
	return &browserk.Config{
		URL:             "http://localhost:8080/",
		AllowedHosts:    []string{"localhost"},
		NumBrowsers:     5,
		MaxDepth:        15,
		FormData:        &browserk.DefaultFormValues,
		DisabledPlugins: nil,
	}
}

// MakeMockAddressForm for an example address form
func MakeMockAddressForm() *browserk.HTMLFormElement {

	children := make([]*browserk.HTMLElement, 0)
	children = append(children, MakeMockLabel("fname", "First Name"))
	children = append(children, MakeMockInput("text", "fname", "John"))
	children = append(children, MakeMockLabel("lname", "Last Name"))
	children = append(children, MakeMockInput("text", "lname", "Doe"))
	children = append(children, MakeMockLabel("email", "E-Mail"))
	children = append(children, MakeMockInput("email", "email", "test@test.com"))
	children = append(children, MakeMockLabel("address", "Address"))
	children = append(children, MakeMockInput("text", "address", ""))
	children = append(children, MakeMockLabel("addr2", "Address Line 2"))
	children = append(children, MakeMockInput("text", "addr2", ""))
	children = append(children, MakeMockLabel("city", "City Name"))
	children = append(children, MakeMockInput("text", "city", ""))
	children = append(children, MakeMockLabel("state", "State"))
	children = append(children, MakeMockInput("text", "state", "CA"))
	children = append(children, MakeMockLabel("zip_code", "ZipCode"))
	children = append(children, MakeMockInput("text", "zip_code", ""))
	children = append(children, MakeMockLabel("country", "Country"))
	children = append(children, MakeMockInput("text", "country", "USA"))

	return &browserk.HTMLFormElement{
		FormType: browserk.FormAddress,
		Events:   nil,
		Attributes: map[string]string{
			"action": "/addAddress",
		},
		Hidden:        false,
		NodeDepth:     3,
		ChildElements: children,
		ID:            nil,
	}
}

func MakeMockInput(inputType, name, placeholder string) *browserk.HTMLElement {
	return &browserk.HTMLElement{
		Type:          browserk.INPUT,
		CustomTagName: "",
		Events:        nil,
		Attributes: map[string]string{
			"name":        name,
			"id":          name,
			"type":        inputType,
			"placeholder": placeholder,
		},
		Hidden:    false,
		NodeDepth: 0,
		ID:        nil,
		Value:     "",
	}
}

func MakeMockLabel(forElement, text string) *browserk.HTMLElement {
	return &browserk.HTMLElement{
		Type:          browserk.LABEL,
		CustomTagName: "",
		Events:        nil,
		Attributes: map[string]string{
			"for": forElement,
		},
		InnerText: text,
		Hidden:    false,
		NodeDepth: 0,
		ID:        nil,
		Value:     "",
	}
}

func MakeMockMessages() []*browserk.HTTPMessage {
	m := make([]*browserk.HTTPMessage, 0)
	for i := 0; i < 3; i++ {
		body := []byte(fmt.Sprintf("this is a body %d", i))
		h := sha1.New()
		h.Write(body)

		message := &browserk.HTTPMessage{
			RequestTime: time.Now(),
			Request: &browserk.HTTPRequest{
				RequestId:   fmt.Sprintf("%d", i+1),
				LoaderId:    fmt.Sprintf("%d", i+1),
				DocumentURL: fmt.Sprintf("http://example.com/%d", i+1),
				Request: &gcdapi.NetworkRequest{
					Url:         fmt.Sprintf("http://example.com/%d", i+1),
					UrlFragment: "",
					Method:      "GET",
					Headers: map[string]interface{}{
						"accept": "text/html",
					},
					PostData:         "",
					HasPostData:      false,
					MixedContentType: "",
					InitialPriority:  "",
					ReferrerPolicy:   "",
					IsLinkPreload:    false,
				},
				Timestamp:        0.0,
				WallTime:         0.0,
				Initiator:        nil,
				RedirectResponse: nil,
			},
			RequestMod:   nil,
			ResponseTime: time.Now().Add(time.Second * 3),
			Response: &browserk.HTTPResponse{
				RequestId: fmt.Sprintf("%d", i+1),
				LoaderId:  fmt.Sprintf("%d", i+1),
				Timestamp: 0.0,
				Type:      "Document",
				Response: &gcdapi.NetworkResponse{
					Url:        fmt.Sprintf("http://example.com/%d", i+1),
					Status:     200,
					StatusText: "OK",
					Headers: map[string]interface{}{
						"content-type": "text/html",
					},
					HeadersText: "",
					MimeType:    "text/html",
					RequestHeaders: map[string]interface{}{
						"": nil,
					},
					RequestHeadersText: "",
					ConnectionReused:   false,
					ConnectionId:       0.0,
					RemoteIPAddress:    "",
					RemotePort:         0,
					FromDiskCache:      false,
					FromServiceWorker:  false,
					FromPrefetchCache:  false,
					EncodedDataLength:  0.0,
					Timing: &gcdapi.NetworkResourceTiming{
						RequestTime:       0.0,
						ProxyStart:        0.0,
						ProxyEnd:          0.0,
						DnsStart:          0.0,
						DnsEnd:            0.0,
						ConnectStart:      0.0,
						ConnectEnd:        0.0,
						SslStart:          0.0,
						SslEnd:            0.0,
						WorkerStart:       0.0,
						WorkerReady:       0.0,
						SendStart:         0.0,
						SendEnd:           0.0,
						PushStart:         0.0,
						PushEnd:           0.0,
						ReceiveHeadersEnd: 0.0,
					},
					Protocol:      "",
					SecurityState: "",
					SecurityDetails: &gcdapi.NetworkSecurityDetails{
						Protocol:                          "",
						KeyExchange:                       "",
						KeyExchangeGroup:                  "",
						Cipher:                            "",
						Mac:                               "",
						CertificateId:                     0,
						SubjectName:                       "",
						SanList:                           nil,
						Issuer:                            "",
						ValidFrom:                         0.0,
						ValidTo:                           0.0,
						SignedCertificateTimestampList:    nil,
						CertificateTransparencyCompliance: "",
					},
				},
				FrameId:  "",
				Body:     body,
				BodyHash: h.Sum(nil),
			},
		}
		m = append(m, message)
	}
	return m
}

func MakeMockCookies() []*browserk.Cookie {
	c := make([]*browserk.Cookie, 0)
	for i := 0; i < 3; i++ {
		c = append(c, &browserk.Cookie{
			Name:         fmt.Sprintf("name%d", i+1),
			Value:        fmt.Sprintf("value%d", i+1),
			Domain:       "",
			Path:         "",
			Expires:      0.0,
			Size:         0,
			HTTPOnly:     true,
			Secure:       true,
			Session:      true,
			SameSite:     "",
			Priority:     "",
			ObservedTime: time.Now(),
		})
	}
	return c
}

func MakeMockConsole() []*browserk.ConsoleEvent {
	c := make([]*browserk.ConsoleEvent, 0)
	for i := 0; i < 3; i++ {
		c = append(c, &browserk.ConsoleEvent{
			Source:   "",
			Level:    "",
			Text:     fmt.Sprintf("name%d", i+1),
			URL:      "",
			Line:     0,
			Column:   0,
			Observed: time.Now().Add(time.Second * time.Duration(i)),
		})
	}
	return c
}

func MakeMockStorage() []*browserk.StorageEvent {
	s := make([]*browserk.StorageEvent, 0)
	for i := 0; i < 3; i++ {
		s = append(s, &browserk.StorageEvent{
			Type:           browserk.StorageAddedEvt,
			IsLocalStorage: false,
			SecurityOrigin: "",
			Key:            fmt.Sprintf("key%d", i+1),
			NewValue:       fmt.Sprintf("value%d", i+1),
			OldValue:       "",
			Observed:       time.Now().Add(time.Second * time.Duration(i)),
		})
	}
	return s
}

func MakeMockResult(id []byte) *browserk.NavigationResult {
	r := &browserk.NavigationResult{
		NavigationID:  id,
		DOM:           "<html>nav result</html>",
		StartURL:      "http://example.com/start" + fmt.Sprintf("%x", id),
		EndURL:        "http://example.com/end",
		MessageCount:  1,
		Messages:      MakeMockMessages(),
		Cookies:       MakeMockCookies(),
		ConsoleEvents: MakeMockConsole(),
		StorageEvents: MakeMockStorage(),
		CausedLoad:    false,
		WasError:      false,
		Errors:        nil,
	}
	r.Hash()
	return r
}

func MakeMockNavi(id []byte) *browserk.Navigation {
	return &browserk.Navigation{
		ID:               id,
		StateUpdatedTime: time.Now(),
		TriggeredBy:      1,
		State:            browserk.NavUnvisited,
		Action: &browserk.Action{
			Type:   browserk.ActLoadURL,
			Input:  nil,
			Result: nil,
		},
	}
}
