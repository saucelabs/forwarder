package pac

import (
	"testing"
)

const pacText = `function FindProxyForURL(url, host) {
  if (
    dnsDomainIs(host, "intranet.domain.com") ||
    shExpMatch(host, "(*.abcdomain.com|abcdomain.com)")
  )
    return "DIRECT";

  if (isPlainHostName(host)) return "DIRECT";
  else return "PROXY 127.0.0.1:8080; PROXY 127.0.0.1:8081; DIRECT";
}
`

func TestNew(t *testing.T) {
	type args struct {
		source      string
		proxiesURIs []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Parser
		wantErr bool
	}{
		{
			name: "Should work",
			args: args{
				source:      pacText,
				proxiesURIs: []string{"http://user:pass@127.0.0.1:8080"},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.source, tt.args.proxiesURIs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			pacProxies, err := got.Find("http://example.com")
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			pacProxy := pacProxies[0]
			pacProxyURI := pacProxy.GetURI().String()
			pacProxyURIExpected := "http://user:pass@127.0.0.1:8080"

			if pacProxyURI != pacProxyURIExpected {
				t.Errorf("pacProxyURI() expected = %v, go %v", pacProxyURI, pacProxyURIExpected)
				return
			}
		})
	}
}
