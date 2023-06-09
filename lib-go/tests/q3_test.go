package tests

// import (
// 	"differentiable/datatypes"
// 	"differentiable/server"
// 	"math/rand"
// 	"net"
// 	"sort"
// 	"testing"
// 	"time"
// )

// // 229.18.218.189,http://YQPuiNNlGS.com,2012-11-17,90.37
// type UserVisit struct {
// 	Ip      string
// 	Dest    string
// 	Time    time.Time
// 	Revenue float64
// }

// // http://FbmMxXIAcZ.com,14.942418037315097
// type Ranking struct {
// 	Url  string
// 	Rank float64
// }

// type Result struct {
// 	Ip           string
// 	TotalRevenue float64
// 	AvgRank      float64
// }

// func genRandomIP() string {
// 	ip := make(net.IP, 4)
// 	rand.Read(ip)
// 	return ip.String()
// }

// func genRandomURL() string {
// 	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
// 	b := make([]byte, 10)
// 	for i := range b {
// 		b[i] = charset[rand.Intn(len(charset))]
// 	}
// 	return "http://" + string(b) + ".com"
// }

// func randomTime(min, max time.Time) time.Time {
// 	minUnix := min.Unix()
// 	maxUnix := max.Unix()
// 	delta := maxUnix - minUnix
// 	sec := rand.Int63n(delta) + minUnix
// 	return time.Unix(sec, 0)
// }

// func genUserVisits(n int, groups int, urls []string) []UserVisit {
// 	ips := []string{}
// 	for i := 0; i < groups; i++ {
// 		ips = append(ips, genRandomIP())
// 	}
// 	rand.Seed(time.Now().UnixNano())
// 	userVisits := make([]UserVisit, n)
// 	start, _ := time.Parse("2006-01-02", "1980-01-01")
// 	end, _ := time.Parse("2006-01-02", "2020-01-01")
// 	for i := 0; i < n; i++ {
// 		userVisits[i] = UserVisit{
// 			Ip:      ips[rand.Intn(groups)],
// 			Dest:    urls[rand.Intn(len(urls))],
// 			Time:    randomTime(start, end),
// 			Revenue: rand.Float64() * 100,
// 		}
// 	}
// 	return userVisits
// }

// func genRankings(n int) ([]Ranking, []string) {
// 	rand.Seed(time.Now().UnixNano())
// 	rankings := make([]Ranking, n)
// 	urls := []string{}
// 	for i := 0; i < n; i++ {
// 		rankings[i] = Ranking{
// 			Url:  genRandomURL(),
// 			Rank: rand.Float64() * 100,
// 		}
// 		urls = append(urls, rankings[i].Url)
// 	}
// 	return rankings, urls
// }

// func TestQ3(t *testing.T) {
// 	server, stop := server.NewStandaloneServer()
// 	agent := server.ForkAgent()
// 	defer stop()

// 	// ranking record: 37.5B, * 16 = 0.6KB
// 	// uv record: 50B, * 8 = 0.4KB
// 	uservisitRecords := 16_000_000
// 	rankingRecords := 8_000_000
// 	groups := 1_000_000

// 	upperTime, _ := time.Parse("2006-01-02", "2000-01-01")

// 	uvs := agent.NewDObj(datatypes.DVEC).(*datatypes.DVec)
// 	ranks, urls := genRankings(rankingRecords)
// 	rMap := map[string]float64{}
// 	for _, r := range ranks {
// 		rMap[r.Url] = r.Rank
// 	}

// 	filterred := uvs.Filter(func(arg any) bool {
// 		uv := arg.(UserVisit)
// 		return uv.Time.Before(upperTime)
// 	})

// 	originalUvs := genUserVisits(uservisitRecords, groups, urls)

// 	ipRankMap := filterred.SumBy(func(arg any) (string, float64) {
// 		uv := arg.(UserVisit)
// 		return uv.Ip, rMap[uv.Dest]
// 	})
// 	ipRevenueMap := filterred.SumBy(func(arg any) (string, float64) {
// 		uv := arg.(UserVisit)
// 		return uv.Ip, uv.Revenue
// 	})
// 	ipClickMap := filterred.SumBy(func(arg any) (string, float64) {
// 		uv := arg.(UserVisit)
// 		return uv.Ip, 1
// 	})

// 	t1 := time.Now()
// 	for _, uv := range originalUvs {
// 		uvs.Push(uv)
// 	}
// 	uvs.End()
// 	t2 := time.Now()

// 	ipRankRes := ipRankMap.Await()
// 	ipRevenueRes := ipRevenueMap.Await()
// 	ipClickRes := ipClickMap.Await()

// 	t3 := time.Now()

// 	combinedMap := []Result{}
// 	for ip, rank := range ipRankRes {
// 		combinedMap = append(combinedMap, Result{
// 			Ip:           ip,
// 			TotalRevenue: ipRevenueRes[ip].(float64),
// 			AvgRank:      rank.(float64) / ipClickRes[ip].(float64),
// 		})
// 	}

// 	sort.Slice(combinedMap, func(i, j int) bool {
// 		return combinedMap[i].TotalRevenue < combinedMap[j].TotalRevenue
// 	})
// 	t4 := time.Now()

// 	t.Errorf("Push: %v\nAwait: %v\nSort: %v\nLen: %v", t2.Sub(t1), t3.Sub(t2), t4.Sub(t3), len(combinedMap))
// }
