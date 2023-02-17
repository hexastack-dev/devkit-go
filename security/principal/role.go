package principal

// Role hold information asociated role name.
type Role struct {
	Name string
}

// Roles represented as map to make it easier to lookup a role.
type Roles map[string]Role

// Get lookup role with given name, return Role and true if found,
// return empty Role and false otherwise.
func (r Roles) Get(name string) (Role, bool) {
	role, found := r[name]
	return role, found
}

// Add Role to Roles using Role.Name as key. Add will replace
// any role with the same name.
func (r Roles) Add(role Role) {
	r[role.Name] = role
}

// Values return slice of Role
func (r Roles) Values() []Role {
	roles := make([]Role, 0)
	for _, v := range r {
		roles = append(roles, v)
	}
	return roles
}
