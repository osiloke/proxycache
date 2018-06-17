package proxy

import (
	"reflect"
	"testing"
)

func Test_replaceHLSUrls(t *testing.T) {
	type args struct {
		hlsRaw         []byte
		proxyServerURL string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "replaceHLSUrls",
			args: args{
				hlsRaw: []byte(`#EXTM3U
#EXT-X-TARGETDURATION:10
#EXT-X-ALLOW-CACHE:YES
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:1
#EXTINF:10.000,
http://35.237.185.199/hls/wGtvsK3hsLMXFAMci3Uq_trd.mp4/seg-1-a1.ts
#EXTINF:10.000,
http://35.237.185.199/hls/wGtvsK3hsLMXFAMci3Uq_trd.mp4/seg-2-a1.ts
#EXTINF:10.000,
http://35.237.185.199/hls/wGtvsK3hsLMXFAMci3Uq_trd.mp4/seg-3-a1.ts
#EXT-X-ENDLIST`),
				proxyServerURL: "http://localhost:7071/",
			},
			want: []byte(`#EXTM3U
#EXT-X-TARGETDURATION:10
#EXT-X-ALLOW-CACHE:YES
#EXT-X-PLAYLIST-TYPE:VOD
#EXT-X-VERSION:3
#EXT-X-MEDIA-SEQUENCE:1
#EXTINF:10.000,
http://localhost:7071/http://35.237.185.199/hls/wGtvsK3hsLMXFAMci3Uq_trd.mp4/seg-1-a1.ts
#EXTINF:10.000,
http://localhost:7071/http://35.237.185.199/hls/wGtvsK3hsLMXFAMci3Uq_trd.mp4/seg-2-a1.ts
#EXTINF:10.000,
http://localhost:7071/http://35.237.185.199/hls/wGtvsK3hsLMXFAMci3Uq_trd.mp4/seg-3-a1.ts
#EXT-X-ENDLIST`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceHLSUrls(tt.args.hlsRaw, tt.args.proxyServerURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("replaceHLSUrls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceHLSUrls() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
