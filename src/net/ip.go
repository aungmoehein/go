// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// IP address manipulations
//
// IPv4 addresses are 4 bytes; IPv6 addresses are 16 bytes.
// An IPv4 address can be converted to an IPv6 address by
// adding a canonical prefix (10 zeros, 2 0xFFs).
// This library accepts either size of byte slice but always
// returns 16-byte addresses.

package net

import "internal/bytealg"

// IP address lengths (bytes).
const (
	IPv4len = 4
	IPv6len = 16
)

// An IP is a single IP address, a slice of bytes.
// Functions in this package accept either 4-byte (IPv4)
// or 16-byte (IPv6) slices as input.
//
// Note that in this documentation, referring to an
// IP address as an IPv4 address or an IPv6 address
// is a semantic property of the address, not just the
// length of the byte slice: a 16-byte slice can still
// be an IPv4 address.
type IP []byte

// An IPMask is a bitmask that can be used to manipulate
// IP addresses for IP addressing and routing.
//
// See type IPNet and func ParseCIDR for details.
type IPMask []byte

// An IPNet represents an IP network.
type IPNet struct {
	IP   IP     // network number
	Mask IPMask // network mask
}

// IPv4 returns the IP address (in 16-byte form) of the
// IPv4 address a.b.c.d.
func IPv4(a, b, c, d byte) IP {
	p := make(IP, IPv6len)
	copy(p, v4InV6Prefix)
	p[12] = a
	p[13] = b
	p[14] = c
	p[15] = d
	return p
}

var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

// IPv4Mask returns the IP mask (in 4-byte form) of the
// IPv4 mask a.b.c.d.
func IPv4Mask(a, b, c, d byte) IPMask {
	p := make(IPMask, IPv4len)
	p[0] = a
	p[1] = b
	p[2] = c
	p[3] = d
	return p
}

// CIDRMask returns an IPMask consisting of 'ones' 1 bits
// followed by 0s up to a total length of 'bits' bits.
// For a mask of this form, CIDRMask is the inverse of IPMask.Size.
func CIDRMask(ones, bits int) IPMask {
	if bits != 8*IPv4len && bits != 8*IPv6len {
		return nil
	}
	if ones < 0 || ones > bits {
		return nil
	}
	return putCIDRMask(nil, ones, bits/8)
}

// putCIDRMask will put an IPMask into `m` consisting of `ones` 1 bits
// followed by 0s up to a total length of `bits' bits if the length
// of m is < `l` bytes and returns the same slice m. If m did not have
// sufficient length for the mask l a new m is returned instead.
func putCIDRMask(m IPMask, ones, bits int) IPMask {
	if len(m) < bits {
		m = make(IPMask, bits)
	}

	n := uint(ones)
	for i := 0; i < bits; i++ {
		if n >= 8 {
			m[i] = 0xff
			n -= 8
			continue
		}
		m[i] = ^byte(0xff >> n)
		n = 0
	}
	return m
}

// Well-known IPv4 addresses
var (
	IPv4bcast     = IPv4(255, 255, 255, 255) // limited broadcast
	IPv4allsys    = IPv4(224, 0, 0, 1)       // all systems
	IPv4allrouter = IPv4(224, 0, 0, 2)       // all routers
	IPv4zero      = IPv4(0, 0, 0, 0)         // all zeros
)

// Well-known IPv6 addresses
var (
	IPv6zero                   = IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	IPv6unspecified            = IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	IPv6loopback               = IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	IPv6interfacelocalallnodes = IP{0xff, 0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	IPv6linklocalallnodes      = IP{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x01}
	IPv6linklocalallrouters    = IP{0xff, 0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x02}
)

// IsUnspecified reports whether ip is an unspecified address, either
// the IPv4 address "0.0.0.0" or the IPv6 address "::".
func (ip IP) IsUnspecified() bool {
	return ip.Equal(IPv4zero) || ip.Equal(IPv6unspecified)
}

// IsLoopback reports whether ip is a loopback address.
func (ip IP) IsLoopback() bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 127
	}
	return ip.Equal(IPv6loopback)
}

// IsMulticast reports whether ip is a multicast address.
func (ip IP) IsMulticast() bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0]&0xf0 == 0xe0
	}
	return len(ip) == IPv6len && ip[0] == 0xff
}

// IsInterfaceLocalMulticast reports whether ip is
// an interface-local multicast address.
func (ip IP) IsInterfaceLocalMulticast() bool {
	return len(ip) == IPv6len && ip[0] == 0xff && ip[1]&0x0f == 0x01
}

// IsLinkLocalMulticast reports whether ip is a link-local
// multicast address.
func (ip IP) IsLinkLocalMulticast() bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 224 && ip4[1] == 0 && ip4[2] == 0
	}
	return len(ip) == IPv6len && ip[0] == 0xff && ip[1]&0x0f == 0x02
}

// IsLinkLocalUnicast reports whether ip is a link-local
// unicast address.
func (ip IP) IsLinkLocalUnicast() bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 169 && ip4[1] == 254
	}
	return len(ip) == IPv6len && ip[0] == 0xfe && ip[1]&0xc0 == 0x80
}

// IsGlobalUnicast reports whether ip is a global unicast
// address.
//
// The identification of global unicast addresses uses address type
// identification as defined in RFC 1122, RFC 4632 and RFC 4291 with
// the exception of IPv4 directed broadcast addresses.
// It returns true even if ip is in IPv4 private address space or
// local IPv6 unicast address space.
func (ip IP) IsGlobalUnicast() bool {
	return (len(ip) == IPv4len || len(ip) == IPv6len) &&
		!ip.Equal(IPv4bcast) &&
		!ip.IsUnspecified() &&
		!ip.IsLoopback() &&
		!ip.IsMulticast() &&
		!ip.IsLinkLocalUnicast()
}

// Is p all zeros?
func isZeros(p IP) bool {
	for i := 0; i < len(p); i++ {
		if p[i] != 0 {
			return false
		}
	}
	return true
}

// To4 converts the IPv4 address ip to a 4-byte representation.
// If ip is not an IPv4 address, To4 returns nil.
func (ip IP) To4() IP {
	if len(ip) == IPv4len {
		return ip
	}
	if len(ip) == IPv6len &&
		isZeros(ip[0:10]) &&
		ip[10] == 0xff &&
		ip[11] == 0xff {
		return ip[12:16]
	}
	return nil
}

// To16 converts the IP address ip to a 16-byte representation.
// If ip is not an IP address (it is the wrong length), To16 returns nil.
func (ip IP) To16() IP {
	if len(ip) == IPv4len {
		return IPv4(ip[0], ip[1], ip[2], ip[3])
	}
	if len(ip) == IPv6len {
		return ip
	}
	return nil
}

// Default route masks for IPv4.
var (
	classAMask = IPv4Mask(0xff, 0, 0, 0)
	classBMask = IPv4Mask(0xff, 0xff, 0, 0)
	classCMask = IPv4Mask(0xff, 0xff, 0xff, 0)
)

// DefaultMask returns the default IP mask for the IP address ip.
// Only IPv4 addresses have default masks; DefaultMask returns
// nil if ip is not a valid IPv4 address.
func (ip IP) DefaultMask() IPMask {
	if ip = ip.To4(); ip == nil {
		return nil
	}
	switch {
	case ip[0] < 0x80:
		return classAMask
	case ip[0] < 0xC0:
		return classBMask
	default:
		return classCMask
	}
}

func allFF(b []byte) bool {
	for _, c := range b {
		if c != 0xff {
			return false
		}
	}
	return true
}

// Mask returns the result of masking the IP address ip with mask.
func (ip IP) Mask(mask IPMask) IP {
	return ipMaskWithBuf(nil, ip, mask)
}

// ipMaskWithBuf implements the IP.Mask method, but containing an
// opitional optional bufer to use for the return value. If the buf is
// too small, a new one is allocated.
func ipMaskWithBuf(buf, ip IP, mask IPMask) []byte {
	if len(mask) == IPv6len && len(ip) == IPv4len && allFF(mask[:12]) {
		mask = mask[12:]
	}
	if len(mask) == IPv4len && len(ip) == IPv6len && bytealg.Equal(ip[:12], v4InV6Prefix) {
		ip = ip[12:]
	}
	n := len(ip)
	if n != len(mask) {
		return nil
	}
	if len(buf) < n {
		buf = make(IP, n)
	}
	for i := 0; i < n; i++ {
		buf[i] = ip[i] & mask[i]
	}
	return buf
}

// ubtoa encodes the string form of the integer v to dst[start:] and
// returns the number of bytes written to dst. The caller must ensure
// that dst has sufficient length.
func ubtoa(dst []byte, start int, v byte) int {
	if v < 10 {
		dst[start] = v + '0'
		return 1
	} else if v < 100 {
		dst[start+1] = v%10 + '0'
		dst[start] = v/10 + '0'
		return 2
	}

	dst[start+2] = v%10 + '0'
	dst[start+1] = (v/10)%10 + '0'
	dst[start] = v/100 + '0'
	return 3
}

// String returns the string form of the IP address ip.
// It returns one of 4 forms:
//   - "<nil>", if ip has length 0
//   - dotted decimal ("192.0.2.1"), if ip is an IPv4 or IP4-mapped IPv6 address
//   - IPv6 ("2001:db8::1"), if ip is a valid IPv6 address
//   - the hexadecimal form of ip, without punctuation, if no other cases apply
func (ip IP) String() string {
	p := ip

	if len(ip) == 0 {
		return "<nil>"
	}

	// If IPv4, use dotted notation.
	if p4 := p.To4(); len(p4) == IPv4len {
		const maxIPv4StringLen = len("255.255.255.255")
		b := make([]byte, maxIPv4StringLen)

		n := ubtoa(b, 0, p4[0])
		b[n] = '.'
		n++

		n += ubtoa(b, n, p4[1])
		b[n] = '.'
		n++

		n += ubtoa(b, n, p4[2])
		b[n] = '.'
		n++

		n += ubtoa(b, n, p4[3])
		return string(b[:n])
	}
	if len(p) != IPv6len {
		return "?" + hexString(ip)
	}

	// Find longest run of zeros.
	e0 := -1
	e1 := -1
	for i := 0; i < IPv6len; i += 2 {
		j := i
		for j < IPv6len && p[j] == 0 && p[j+1] == 0 {
			j += 2
		}
		if j > i && j-i > e1-e0 {
			e0 = i
			e1 = j
			i = j
		}
	}
	// The symbol "::" MUST NOT be used to shorten just one 16 bit 0 field.
	if e1-e0 <= 2 {
		e0 = -1
		e1 = -1
	}

	const maxLen = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	b := make([]byte, 0, maxLen)

	// Print with possible :: in place of run of zeros
	for i := 0; i < IPv6len; i += 2 {
		if i == e0 {
			b = append(b, ':', ':')
			i = e1
			if i >= IPv6len {
				break
			}
		} else if i > 0 {
			b = append(b, ':')
		}
		b = appendHex(b, (uint32(p[i])<<8)|uint32(p[i+1]))
	}
	return string(b)
}

func hexString(b []byte) string {
	s := make([]byte, len(b)*2)
	for i, tn := range b {
		s[i*2], s[i*2+1] = hexDigit[tn>>4], hexDigit[tn&0xf]
	}
	return string(s)
}

// ipEmptyString is like ip.String except that it returns
// an empty string when ip is unset.
func ipEmptyString(ip IP) string {
	if len(ip) == 0 {
		return ""
	}
	return ip.String()
}

// MarshalText implements the encoding.TextMarshaler interface.
// The encoding is the same as returned by String, with one exception:
// When len(ip) is zero, it returns an empty slice.
func (ip IP) MarshalText() ([]byte, error) {
	if len(ip) == 0 {
		return []byte(""), nil
	}
	if len(ip) != IPv4len && len(ip) != IPv6len {
		return nil, &AddrError{Err: "invalid IP address", Addr: hexString(ip)}
	}
	return []byte(ip.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The IP address is expected in a form accepted by ParseIP.
func (ip *IP) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*ip = nil
		return nil
	}
	s := string(text)
	x := ParseIP(s)
	if x == nil {
		return &ParseError{Type: "IP address", Text: s}
	}
	*ip = x
	return nil
}

// Equal reports whether ip and x are the same IP address.
// An IPv4 address and that same address in IPv6 form are
// considered to be equal.
func (ip IP) Equal(x IP) bool {
	if len(ip) == len(x) {
		return bytealg.Equal(ip, x)
	}
	if len(ip) == IPv4len && len(x) == IPv6len {
		return bytealg.Equal(x[0:12], v4InV6Prefix) && bytealg.Equal(ip, x[12:])
	}
	if len(ip) == IPv6len && len(x) == IPv4len {
		return bytealg.Equal(ip[0:12], v4InV6Prefix) && bytealg.Equal(ip[12:], x)
	}
	return false
}

func (ip IP) matchAddrFamily(x IP) bool {
	return ip.To4() != nil && x.To4() != nil || ip.To16() != nil && ip.To4() == nil && x.To16() != nil && x.To4() == nil
}

// If mask is a sequence of 1 bits followed by 0 bits,
// return the number of 1 bits.
func simpleMaskLength(mask IPMask) int {
	var n int
	for i, v := range mask {
		if v == 0xff {
			n += 8
			continue
		}
		// found non-ff byte
		// count 1 bits
		for v&0x80 != 0 {
			n++
			v <<= 1
		}
		// rest must be 0 bits
		if v != 0 {
			return -1
		}
		for i++; i < len(mask); i++ {
			if mask[i] != 0 {
				return -1
			}
		}
		break
	}
	return n
}

// Size returns the number of leading ones and total bits in the mask.
// If the mask is not in the canonical form--ones followed by zeros--then
// Size returns 0, 0.
func (m IPMask) Size() (ones, bits int) {
	ones, bits = simpleMaskLength(m), len(m)*8
	if ones == -1 {
		return 0, 0
	}
	return
}

// String returns the hexadecimal form of m, with no punctuation.
func (m IPMask) String() string {
	if len(m) == 0 {
		return "<nil>"
	}
	return hexString(m)
}

func networkNumberAndMask(n *IPNet) (ip IP, m IPMask) {
	if ip = n.IP.To4(); ip == nil {
		ip = n.IP
		if len(ip) != IPv6len {
			return nil, nil
		}
	}
	m = n.Mask
	switch len(m) {
	case IPv4len:
		if len(ip) != IPv4len {
			return nil, nil
		}
	case IPv6len:
		if len(ip) == IPv4len {
			m = m[12:]
		}
	default:
		return nil, nil
	}
	return
}

// Contains reports whether the network includes ip.
func (n *IPNet) Contains(ip IP) bool {
	nn, m := networkNumberAndMask(n)
	if x := ip.To4(); x != nil {
		ip = x
	}
	l := len(ip)
	if l != len(nn) {
		return false
	}
	for i := 0; i < l; i++ {
		if nn[i]&m[i] != ip[i]&m[i] {
			return false
		}
	}
	return true
}

// Network returns the address's network name, "ip+net".
func (n *IPNet) Network() string { return "ip+net" }

// String returns the CIDR notation of n like "192.0.2.0/24"
// or "2001:db8::/48" as defined in RFC 4632 and RFC 4291.
// If the mask is not in the canonical form, it returns the
// string which consists of an IP address, followed by a slash
// character and a mask expressed as hexadecimal form with no
// punctuation like "198.51.100.0/c000ff00".
func (n *IPNet) String() string {
	nn, m := networkNumberAndMask(n)
	if nn == nil || m == nil {
		return "<nil>"
	}
	l := simpleMaskLength(m)
	if l == -1 {
		return nn.String() + "/" + m.String()
	}
	return nn.String() + "/" + uitoa(uint(l))
}

// Parse IPv4 address (d.d.d.d).
func parseIPv4(s string) IP {
	var p [IPv4len]byte
	if !parseIntoIPv4(p[:], s) {
		return nil
	}
	return IPv4(p[0], p[1], p[2], p[3])
}

// parseIntoIPv4 parses s as an IPv4 address and parses it into
// p, which must be at least IPv4len bytes long. It reports
// whether the parse was successful.
func parseIntoIPv4(p []byte, s string) bool {
	for i := 0; i < IPv4len; i++ {
		if len(s) == 0 {
			// Missing octets.
			return false
		}
		if i > 0 {
			if s[0] != '.' {
				return false
			}
			s = s[1:]
		}
		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return false
		}
		s = s[c:]
		p[i] = byte(n)
	}
	return len(s) == 0
}

// parseIPv6Zone parses s as a literal IPv6 address and its associated zone
// identifier which is described in RFC 4007.
func parseIPv6Zone(s string) (IP, string) {
	s, zone := splitHostZone(s)
	return parseIPv6(s), zone
}

// parseIPv6 parses s as a literal IPv6 address described in RFC 4291
// and RFC 5952.
func parseIPv6(s string) (ip IP) {
	if ip = make(IP, IPv6len); !parseIntoIPv6(ip, s) {
		ip = nil
	}
	return
}

// parseIntoIPv6 parses s as a literal IPv6 address described in RFC
// 4291 and RFC 5952 and populates ip. It reports whether the parse
// was successful.
func parseIntoIPv6(ip IP, s string) bool {
	ellipsis := -1 // position of ellipsis in ip

	// Might have leading ellipsis
	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		// Might be only ellipsis
		if len(s) == 0 {
			return true
		}
	}

	// Loop, parsing hex numbers followed by colon.
	i := 0
	for i < IPv6len {
		// Hex number.
		n, c, ok := xtoi(s)
		if !ok || n > 0xFFFF {
			return false
		}

		// If followed by dot, might be in trailing IPv4.
		if c < len(s) && s[c] == '.' {
			if ellipsis < 0 && i != IPv6len-IPv4len {
				// Not the right place.
				return false
			}
			if i+IPv4len > IPv6len {
				// Not enough room.
				return false
			}
			ip4 := parseIPv4(s)
			if ip4 == nil {
				return false
			}
			ip[i] = ip4[12]
			ip[i+1] = ip4[13]
			ip[i+2] = ip4[14]
			ip[i+3] = ip4[15]
			s = ""
			i += IPv4len
			break
		}

		// Save this 16-bit chunk.
		ip[i] = byte(n >> 8)
		ip[i+1] = byte(n)
		i += 2

		// Stop at end of string.
		s = s[c:]
		if len(s) == 0 {
			break
		}

		// Otherwise must be followed by colon and more.
		if s[0] != ':' || len(s) == 1 {
			return false
		}
		s = s[1:]

		// Look for ellipsis.
		if s[0] == ':' {
			if ellipsis >= 0 { // already have one
				return false
			}
			ellipsis = i
			s = s[1:]
			if len(s) == 0 { // can be at end
				break
			}
		}
	}

	// Must have used entire string.
	if len(s) != 0 {
		return false
	}

	// If didn't parse enough, expand ellipsis.
	if i < IPv6len {
		if ellipsis < 0 {
			return false
		}
		n := IPv6len - i
		for j := i - 1; j >= ellipsis; j-- {
			ip[j+n] = ip[j]
		}
		for j := ellipsis + n - 1; j >= ellipsis; j-- {
			ip[j] = 0
		}
	} else if ellipsis >= 0 {
		// Ellipsis must represent at least one 0 group.
		return false
	}
	return true
}

// ParseIP parses s as an IP address, returning the result.
// The string s can be in dotted decimal ("192.0.2.1")
// or IPv6 ("2001:db8::68") form.
// If s is not a valid textual representation of an IP address,
// ParseIP returns nil.
func ParseIP(s string) IP {
	switch n := strToIPvLen(s); n {
	case IPv4len:
		return parseIPv4(s)
	case IPv6len:
		return parseIPv6(s)
	default:
		return nil
	}
}

// strToIPvLen reports the expected length of the parsed IP adress
// given its ASCII representation in s. It returns IPv4len, IPv6len,
// or -1.
func strToIPvLen(s string) int {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return IPv4len
		case ':':
			return IPv6len
		}
	}
	return -1
}

// parseIPZone parses s as an IP address, return it and its associated zone
// identifier (IPv6 only).
func parseIPZone(s string) (IP, string) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return parseIPv4(s), ""
		case ':':
			return parseIPv6Zone(s)
		}
	}
	return nil, ""
}

// ParseCIDR parses s as a CIDR notation IP address and prefix length,
// like "192.0.2.0/24" or "2001:db8::/32", as defined in
// RFC 4632 and RFC 4291.
//
// It returns the IP address and the network implied by the IP and
// prefix length.
// For example, ParseCIDR("192.0.2.1/24") returns the IP address
// 192.0.2.1 and the network 192.0.2.0/24.
func ParseCIDR(s string) (IP, *IPNet, error) {
	i := bytealg.IndexByteString(s, '/')
	if i < 0 {
		return nil, nil, &ParseError{Type: "CIDR address", Text: s}
	}

	iplen := strToIPvLen(s[:i])
	if iplen < 0 {
		return nil, nil, &ParseError{Type: "CIDR address", Text: s}
	}

	nwlen, y, ok := dtoi(s[i+1:])
	if !ok || y != len(s[i+1:]) || nwlen < 0 || nwlen > 8*iplen {
		return nil, nil, &ParseError{Type: "CIDR address", Text: s}
	}

	var (
		buf []byte
		ipn *IPNet
	)
	switch iplen {
	case IPv4len:
		var block struct {
			ipn IPNet
			arr [IPv6len + (IPv4len * 2)]byte
		}
		ipn = &block.ipn
		buf = block.arr[:]
		copy(buf, v4InV6Prefix)
		if !parseIntoIPv4(buf[len(v4InV6Prefix):], s[:i]) {
			return nil, nil, &ParseError{Type: "CIDR address", Text: s}
		}
	case IPv6len:
		var block struct {
			ipn IPNet
			arr [IPv6len * 3]byte
		}
		ipn = &block.ipn
		buf = block.arr[:]
		if !parseIntoIPv6(buf, s[:i]) {
			return nil, nil, &ParseError{Type: "CIDR address", Text: s}
		}
	default:
		return nil, nil, &ParseError{Type: "CIDR address", Text: s}
	}

	ip := buf[:IPv6len:IPv6len]
	buf = buf[IPv6len:]

	ipn.Mask = putCIDRMask(buf[:iplen:iplen], nwlen, iplen)
	buf = buf[iplen:]

	ipn.IP = ipMaskWithBuf(buf[:iplen:iplen], ip, ipn.Mask)
	return ip, ipn, nil
}
