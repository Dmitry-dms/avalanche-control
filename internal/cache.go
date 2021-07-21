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
	Users       []ClientStat 
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
	stat := &stats{
		OnlineUsers: info.OnlineUsers,
		MaxUsers: info.MaxUsers,
		Users: info.Users,
	}
	c.Info[companyName] = stat
}
