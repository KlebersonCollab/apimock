package template

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	firstNames = []string{
		"Alex", "Emma", "Liam", "Olivia", "Noah", "Ava", "Lucas", "Sophia", "Ethan", "Isabella",
		"Mason", "Mia", "Oliver", "Charlotte", "Elijah", "Amelia", "Mateo", "Harper", "Logan", "Evelyn",
		"Carlos", "Sofia", "Gabriel", "Julia", "Arthur", "Helena", "Bernardo", "Alice", "Heitor", "Laura",
		"Kenji", "Yuki", "Hana", "Ren", "Aoi", "Dmitri", "Elena", "Sven", "Astrid", "Jean", "Camille",
	}

	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
		"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
		"Silva", "Santos", "Oliveira", "Souza", "Pereira", "Lima", "Carvalho", "Ferreira", "Ribeiro", "Almeida",
		"Tanaka", "Sato", "Suzuki", "Takahashi", "Watanabe", "Muller", "Schmidt", "Dubois", "Moreau", "Novak",
	}

	domains = []string{
		"gmail.com", "outlook.com", "yahoo.com", "icloud.com", "proton.me", "linear.app", "github.com", "stripe.com", "vercel.app", "acme.corp",
	}

	jobTitles = []string{
		"Senior Software Engineer", "Product Designer", "Staff Frontend Architect", "Engineering Manager",
		"VP of Product", "Data Scientist", "DevOps Specialist", "UX Researcher", "Security Engineer",
		"Full Stack Developer", "Systems Architect", "QA Automation Lead", "Cloud Engineer", "CTO",
	}

	companies = []string{
		"Acme Corp", "Apex Labs", "Starlight Systems", "Nexus Technologies", "Pulse Dynamics",
		"Horizon Media", "Vortex Digital", "Synthetix AI", "OmniFlow Cloud", "HyperScale Logic",
		"Aether Dynamics", "CipherKey Security", "Nova Solutions", "Quantum Leap Tech", "Linear Flow",
	}

	cities = []string{
		"San Francisco", "New York", "London", "Tokyo", "Berlin", "Sao Paulo", "Paris", "Toronto",
		"Amsterdam", "Singapore", "Sydney", "Stockholm", "Austin", "Zurich", "Dublin", "Seoul",
	}

	countries = []string{
		"United States", "United Kingdom", "Germany", "Japan", "Brazil", "Canada", "France",
		"Netherlands", "Australia", "Singapore", "Sweden", "Switzerland", "Ireland", "South Korea",
	}

	streetNames = []string{
		"Market Street", "Broadway", "Oxford Street", "Shibuya Crossing", "Kurfurstendamm",
		"Avenida Paulista", "Champs-Elysees", "Bay Street", "George Street", "Silicon Avenue",
	}

	loremWords = []string{
		"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit", "sed", "do",
		"eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua", "enim",
		"ad", "minim", "veniam", "quis", "nostrud", "exercitation", "ullamco", "laboris", "nisi",
		"ut", "aliquip", "ex", "ea", "commodo", "consequat", "duis", "aute", "irure", "in",
		"reprehenderit", "voluptate", "velit", "esse", "cillum", "fugiat", "nulla", "pariatur",
		"excepteur", "sint", "occaecat", "cupidatat", "non", "proident", "sunt", "culpa", "qui",
		"officia", "deserunt", "mollit", "anim", "id", "est", "laborum", "resilience", "latency",
		"performance", "scalable", "distributed", "interface", "telemetry", "deterministic", "concurrency",
	}

	statuses = []string{
		"active", "pending", "completed", "archived", "suspended", "in_progress", "draft", "published",
	}

	roles = []string{
		"admin", "developer", "viewer", "editor", "owner", "billing_admin", "guest",
	}

	currencies = []string{"USD", "EUR", "GBP", "BRL", "JPY", "CAD", "AUD"}

	categories = []string{
		"technology", "electronics", "fashion", "books", "home", "fitness", "gaming", "music",
	}
)

// Faker provides realistic data generation methods
type Faker struct {
	mu sync.Mutex
}

// NewFaker initializes a Faker instance
func NewFaker() *Faker {
	return &Faker{}
}

func (f *Faker) randomInt(min, max int) int {
	if min >= max {
		return min
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return min + int(nBig.Int64())
}

func (f *Faker) randomChoice(list []string) string {
	if len(list) == 0 {
		return ""
	}
	idx := f.randomInt(0, len(list)-1)
	return list[idx]
}

// FirstName returns a realistic first name
func (f *Faker) FirstName() string {
	return f.randomChoice(firstNames)
}

// LastName returns a realistic last name
func (f *Faker) LastName() string {
	return f.randomChoice(lastNames)
}

// Name returns a full name
func (f *Faker) Name() string {
	return fmt.Sprintf("%s %s", f.FirstName(), f.LastName())
}

// Username generates a clean username
func (f *Faker) Username() string {
	first := strings.ToLower(f.FirstName())
	last := strings.ToLower(f.LastName())
	num := f.randomInt(10, 999)
	return fmt.Sprintf("%s.%s%d", first, last, num)
}

// Email returns a realistic email
func (f *Faker) Email() string {
	first := strings.ToLower(f.FirstName())
	last := strings.ToLower(f.LastName())
	domain := f.randomChoice(domains)
	return fmt.Sprintf("%s.%s@%s", first, last, domain)
}

// Avatar returns an avatar URL
func (f *Faker) Avatar() string {
	id := f.randomInt(1, 70)
	gender := "men"
	if f.randomInt(0, 1) == 1 {
		gender = "women"
	}
	return fmt.Sprintf("https://randomuser.me/api/portraits/%s/%d.jpg", gender, id)
}

// Phone returns a formatted telephone number
func (f *Faker) Phone() string {
	return fmt.Sprintf("+1 (%d) %d-%d", f.randomInt(200, 999), f.randomInt(100, 999), f.randomInt(1000, 9999))
}

// JobTitle returns a professional job title
func (f *Faker) JobTitle() string {
	return f.randomChoice(jobTitles)
}

// Company returns a company name
func (f *Faker) Company() string {
	return f.randomChoice(companies)
}

// Street returns a street address
func (f *Faker) Street() string {
	return fmt.Sprintf("%d %s", f.randomInt(10, 9999), f.randomChoice(streetNames))
}

// City returns a city name
func (f *Faker) City() string {
	return f.randomChoice(cities)
}

// Country returns a country name
func (f *Faker) Country() string {
	return f.randomChoice(countries)
}

// ZipCode returns a postal code
func (f *Faker) ZipCode() string {
	return fmt.Sprintf("%05d-%04d", f.randomInt(10000, 99999), f.randomInt(1000, 9999))
}

// UUID returns a standard v4 UUID string
func (f *Faker) UUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ID returns an incremental or random integer ID
func (f *Faker) ID() int {
	return f.randomInt(1, 10000)
}

// Number returns an integer between min and max
func (f *Faker) Number(min, max int) int {
	return f.randomInt(min, max)
}

// Float returns a float between min and max formatted with decimals
func (f *Faker) Float(min, max float64, decimals int) float64 {
	if min >= max {
		return min
	}
	r := float64(f.randomInt(0, 1000000)) / 1000000.0
	val := min + r*(max-min)
	format := fmt.Sprintf("%%.%df", decimals)
	formatted, _ := strconv.ParseFloat(fmt.Sprintf(format, val), 64)
	return formatted
}

// Boolean returns a boolean
func (f *Faker) Boolean() bool {
	return f.randomInt(0, 1) == 1
}

// Price returns a formatted price float
func (f *Faker) Price(min, max float64) float64 {
	return f.Float(min, max, 2)
}

// Currency returns a 3-letter currency code
func (f *Faker) Currency() string {
	return f.randomChoice(currencies)
}

// Status returns a typical entity status
func (f *Faker) Status() string {
	return f.randomChoice(statuses)
}

// Role returns a user authorization role
func (f *Faker) Role() string {
	return f.randomChoice(roles)
}

// Lorem generates words of lorem ipsum
func (f *Faker) Lorem(wordCount int) string {
	if wordCount <= 0 {
		wordCount = 5
	}
	var res []string
	for i := 0; i < wordCount; i++ {
		res = append(res, f.randomChoice(loremWords))
	}
	return strings.Join(res, " ")
}

// Sentence generates a capitalized sentence
func (f *Faker) Sentence() string {
	words := f.Lorem(f.randomInt(6, 14))
	return strings.ToUpper(words[:1]) + words[1:] + "."
}

// Paragraph generates a multi-sentence paragraph
func (f *Faker) Paragraph() string {
	sentences := make([]string, f.randomInt(3, 5))
	for i := range sentences {
		sentences[i] = f.Sentence()
	}
	return strings.Join(sentences, " ")
}

// Date returns a date formatted YYYY-MM-DD
func (f *Faker) Date() string {
	return time.Now().AddDate(0, 0, -f.randomInt(0, 365)).Format("2006-01-02")
}

// DateTime returns ISO8601 string
func (f *Faker) DateTime() string {
	return time.Now().Add(time.Duration(-f.randomInt(0, 86400*30)) * time.Second).Format(time.RFC3339)
}

// PastDate returns a date N days in the past
func (f *Faker) PastDate(days int) string {
	if days <= 0 {
		days = 30
	}
	return time.Now().AddDate(0, 0, -f.randomInt(1, days)).Format("2006-01-02")
}

// FutureDate returns a date N days in the future
func (f *Faker) FutureDate(days int) string {
	if days <= 0 {
		days = 30
	}
	return time.Now().AddDate(0, 0, f.randomInt(1, days)).Format("2006-01-02")
}

// Image returns a placeholder image URL
func (f *Faker) Image(width, height int, category string) string {
	if width <= 0 {
		width = 600
	}
	if height <= 0 {
		height = 400
	}
	if category == "" {
		category = f.randomChoice(categories)
	}
	return fmt.Sprintf("https://picsum.photos/%d/%d?category=%s", width, height, category)
}
