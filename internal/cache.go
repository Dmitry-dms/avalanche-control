package internal

type CompanyStats struct {
	Name        string       `json:"company_name"`
	OnlineUsers uint         `json:"online_users"`
	MaxUsers    uint         `json:"max_users"`
	Users       []ClientStat `json:"active_users"`
}
type stats struct {
	OnlineUsers uint
	MaxUsers    uint
	Users       map[string]struct{}
}
type ClientStat struct {
	UserId string `json:"user_id"`
}

type Cache struct {
	Info map[string]*stats
}

func NewCache() *Cache {
	return &Cache{
		Info: make(map[string]*stats),
	}
}
func (c *Cache) Show() map[string]*stats {
	return c.Info
}
func (c *Cache) update(companyName string, info *CompanyStats) {
	users := make(map[string]struct{})
	for i := range info.Users {
		users[info.Users[i].UserId] = struct{}{}
	}
	stat := &stats{
		OnlineUsers: info.OnlineUsers,
		MaxUsers:    info.MaxUsers,
		Users:       users,
	}
	c.Info[companyName] = stat
}
func (c *Cache) GetUser(companyName string, userId string) (bool, bool) {
	company, ok := c.Info[companyName]
	_, u := company.Users[userId]
	return ok, u
}
func (c *Cache) GetCompany(companyName string) bool {
	_, ok := c.Info[companyName]
	return ok
}