package vpnid

import (
	"net"
	"reflect"
	"testing"
)

func mustIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		panic("bad IP: " + s)
	}
	return ip.To4()
}

func TestCollapse(t *testing.T) {
	tests := []struct {
		name string
		ips  []string
		want []string
	}{
		{
			name: "empty input",
			ips:  []string{},
			want: []string{},
		},
		{
			name: "single IP",
			ips:  []string{"192.168.0.1"},
			want: []string{"192.168.0.1/32"},
		},
		{
			name: "two non-consecutive IPs",
			ips:  []string{"192.168.0.1", "192.168.0.3"},
			want: []string{"192.168.0.1/32", "192.168.0.3/32"},
		},
		{
			name: "four consecutive IPs form /30",
			ips:  []string{"192.168.0.1", "192.168.0.2", "192.168.0.3", "192.168.0.4"},
			want: []string{"192.168.0.0/30"},
		},
		{
			name: "consecutive then gap",
			ips:  []string{"10.0.0.1", "10.0.0.2", "192.168.0.1"},
			want: []string{"10.0.0.1/32", "10.0.0.2/32", "192.168.0.1/32"},
		},
		{
			name: "large block /29",
			ips: []string{
				"192.168.1.0",
				"192.168.1.1",
				"192.168.1.2",
				"192.168.1.3",
				"192.168.1.4",
				"192.168.1.5",
				"192.168.1.6",
				"192.168.1.7",
			},
			want: []string{"192.168.1.0/29"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ips []net.IP
			for _, ip := range tt.ips {
				ips = append(ips, mustIP(ip))
			}

			gotNets := Collapse(ips)

			// normalize both got/want to []string to cover empty return case
			var got []string
			for _, n := range gotNets {
				got = append(got, n.String())
			}

			if len(tt.want) != 0 && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collapse(%v) = %v, want %v", tt.ips, got, tt.want)
			} else if len(tt.want) == 0 { // to cover empty response never fulfilling deep equal
				if len(got) != 0 {
					t.Errorf("collapse(%v) = %v, want %v", tt.ips, got, nil)
				}
			}
		})
	}
}
