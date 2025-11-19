package services

import (
	"strings"
)

// SkillSynonyms maps skills to their synonyms and related skills
var SkillSynonyms = map[string][]string{
	// Microsoft Office Suite - Enhanced recognition
	"excel":           {"microsoft office", "ms office", "office suite", "spreadsheet", "microsoft excel", "excel spreadsheet", "ms excel", "office excel", "excel 365", "excel 2019", "excel 2016"},
	"word":            {"microsoft office", "ms office", "office suite", "microsoft word", "ms word", "office word", "word 365", "word 2019"},
	"powerpoint":      {"microsoft office", "ms office", "office suite", "presentation", "microsoft powerpoint", "ms powerpoint", "ppt", "powerpoint 365"},
	"microsoft office": {"excel", "word", "powerpoint", "ms office", "office suite", "outlook", "access", "office 365", "office 2019", "office 2016", "msoffice"},
	"ms office":       {"excel", "word", "powerpoint", "microsoft office", "office suite", "office 365", "msoffice"},
	"office suite":    {"microsoft office", "ms office", "excel", "word", "powerpoint", "office 365"},
	
	// Programming Languages - Variations
	"react":           {"react.js", "reactjs", "react js", "react native"},
	"react.js":        {"react", "reactjs", "react js"},
	"reactjs":         {"react", "react.js", "react js"},
	"node.js":         {"nodejs", "node", "node js"},
	"nodejs":          {"node.js", "node", "node js"},
	"node":            {"node.js", "nodejs", "node js"},
	"javascript":     {"js", "ecmascript", "es6", "es7", "es8", "typescript"},
	"js":              {"javascript", "ecmascript", "es6", "es7", "es8"},
	"typescript":      {"ts", "javascript", "js"},
	"ts":              {"typescript"},
	"c++":             {"cpp", "c plus plus", "cplusplus"},
	"c#":              {"csharp", "c sharp", "dotnet", ".net"},
	".net":            {"dotnet", "c#", "csharp", "asp.net"},
	"python":          {"py", "python3", "python 3"},
	
	// Frameworks & Libraries
	"angular":         {"angularjs", "angular.js", "angular 2", "angular 2+"},
	"angular.js":      {"angular", "angularjs", "angular 2", "angular 2+"},
	"vue":             {"vue.js", "vuejs", "vue js", "vue 3"},
	"vue.js":          {"vue", "vuejs", "vue js", "vue 3"},
	"express":         {"express.js", "expressjs", "express js"},
	
	// Databases
	"postgresql":      {"postgres", "pg"},
	"postgres":        {"postgresql", "pg"},
	"mongodb":         {"mongo", "mongo db", "nosql"},
	"mongo":           {"mongodb", "mongo db", "nosql"},
	"nosql":           {"mongodb", "mongo", "mongo db"},
	"mysql":           {"mariadb", "sql"},
	"sql":             {"sql server", "mysql", "postgresql", "postgres", "database"},
	
	// Cloud & DevOps
	"aws":             {"amazon web services", "amazon aws", "ec2", "s3", "lambda", "amazon cloud"},
	"docker":          {"containerization", "containers", "dockerfile", "docker containers"},
	"kubernetes":      {"k8s", "kube", "container orchestration"},
	"k8s":             {"kubernetes", "kube", "container orchestration"},
	"git":             {"github", "gitlab", "version control", "scm", "source control", "git version control"},
	
	// Methodologies
	"agile":           {"scrum", "kanban", "sprint", "agile methodology", "agile development"},
	"scrum":           {"agile", "sprint", "scrum master", "scrum methodology"},
	
	// Web Technologies
	"html":            {"html5", "hypertext markup language"},
	"html5":           {"html", "hypertext markup language"},
	"css":             {"css3", "stylesheet", "styling"},
	"css3":            {"css", "stylesheet", "styling"},
	"rest":            {"rest api", "restful", "restful api"},
	"rest api":        {"rest", "restful", "restful api"},
	"api":             {"rest", "rest api", "graphql", "web api", "apis"},
	
	// Design Tools
	"photoshop":       {"adobe photoshop", "ps", "adobe creative suite"},
	"illustrator":    {"adobe illustrator", "ai", "adobe creative suite"},
	"figma":           {"ui design", "ux design", "design tool"},
}

// JobTitleSkillInference maps job titles to inferred skills
// Enhanced with better inference - Senior roles imply basic skills
var JobTitleSkillInference = map[string][]string{
	// Developer roles - Enhanced inference
	"developer":           {"programming", "coding", "software development", "problem solving", "git", "debugging", "algorithms", "data structures"},
	"senior developer":    {"programming", "coding", "software development", "problem solving", "git", "debugging", "architecture", "mentoring", "code review", "algorithms", "data structures", "system design", "best practices"},
	"software developer": {"programming", "coding", "software development", "problem solving", "git", "debugging", "algorithms", "data structures"},
	"full stack developer": {"javascript", "html", "css", "database", "api", "frontend", "backend", "full stack", "programming", "coding", "git"},
	"frontend developer":  {"html", "css", "javascript", "ui", "ux", "responsive design", "frontend", "programming", "coding", "git"},
	"backend developer":   {"api", "database", "server", "backend", "rest", "sql", "programming", "coding", "git"},
	"web developer":       {"html", "css", "javascript", "web development", "responsive design", "programming", "coding", "git"},
	"junior developer":     {"programming", "coding", "software development", "problem solving", "git", "debugging", "learning"},
	"mid-level developer":  {"programming", "coding", "software development", "problem solving", "git", "debugging", "algorithms", "data structures"},
	"lead developer":      {"programming", "coding", "software development", "problem solving", "git", "debugging", "architecture", "mentoring", "code review", "leadership", "system design"},
	
	// Engineer roles - Enhanced inference
	"software engineer":    {"programming", "coding", "software development", "problem solving", "git", "debugging", "algorithms", "data structures", "system design"},
	"senior engineer":      {"programming", "coding", "software development", "problem solving", "git", "debugging", "architecture", "mentoring", "code review", "system design", "algorithms", "data structures"},
	"senior software engineer": {"programming", "coding", "software development", "problem solving", "git", "debugging", "architecture", "mentoring", "code review", "system design", "algorithms", "data structures", "best practices"},
	"devops engineer":      {"docker", "kubernetes", "ci/cd", "cloud", "aws", "infrastructure", "automation", "scripting", "linux"},
	"qa engineer":          {"testing", "quality assurance", "test automation", "bug tracking", "test cases"},
	"qa":                   {"testing", "quality assurance", "test automation", "bug tracking", "test cases"},
	
	// Manager roles
	"project manager":      {"project management", "agile", "scrum", "planning", "coordination", "communication"},
	"product manager":      {"product management", "strategy", "planning", "communication", "analytics"},
	"team lead":            {"leadership", "mentoring", "code review", "planning", "coordination"},
	
	// Data roles
	"data analyst":         {"sql", "excel", "data analysis", "analytics", "reporting"},
	"data scientist":      {"python", "sql", "machine learning", "data analysis", "statistics"},
	
	// Designer roles
	"ui designer":         {"ui design", "figma", "photoshop", "illustrator", "design"},
	"ux designer":         {"ux design", "user research", "wireframing", "prototyping", "figma"},
	"graphic designer":    {"photoshop", "illustrator", "design", "creative", "adobe creative suite"},
}

// GetSkillSynonyms returns all synonyms and related skills for a given skill
func GetSkillSynonyms(skill string) []string {
	skillLower := strings.ToLower(strings.TrimSpace(skill))
	
	// Direct match
	if synonyms, ok := SkillSynonyms[skillLower]; ok {
		return synonyms
	}
	
	// Reverse lookup - find if this skill is a synonym of another
	for mainSkill, synonyms := range SkillSynonyms {
		for _, synonym := range synonyms {
			if strings.EqualFold(synonym, skillLower) {
				// Return the main skill and all its synonyms
				result := []string{mainSkill}
				result = append(result, synonyms...)
				return result
			}
		}
	}
	
	return []string{skillLower}
}

// InferSkillsFromJobTitle infers skills based on job title
func InferSkillsFromJobTitle(jobTitle string) []string {
	titleLower := strings.ToLower(jobTitle)
	inferredSkills := []string{}
	
	// Check for exact matches
	if skills, ok := JobTitleSkillInference[titleLower]; ok {
		return skills
	}
	
	// Check for partial matches (e.g., "Senior Software Developer" contains "developer")
	for titleKey, skills := range JobTitleSkillInference {
		if strings.Contains(titleLower, titleKey) {
			inferredSkills = append(inferredSkills, skills...)
		}
	}
	
	// Remove duplicates
	uniqueSkills := make(map[string]bool)
	result := []string{}
	for _, skill := range inferredSkills {
		if !uniqueSkills[skill] {
			uniqueSkills[skill] = true
			result = append(result, skill)
		}
	}
	
	return result
}

// NormalizeSkill normalizes a skill name for matching
func NormalizeSkill(skill string) string {
	skill = strings.ToLower(strings.TrimSpace(skill))
	
	// Remove common variations
	skill = strings.ReplaceAll(skill, ".", "")
	skill = strings.ReplaceAll(skill, "-", " ")
	skill = strings.ReplaceAll(skill, "_", " ")
	skill = strings.ReplaceAll(skill, "+", "plus")
	
	// Remove extra spaces
	words := strings.Fields(skill)
	return strings.Join(words, " ")
}

