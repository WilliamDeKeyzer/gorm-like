package gormlike

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ing-bank/gormtestutil"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type ObjectA struct {
	ID    uuid.UUID `gorm:"primaryKey"`
	Name  string
	Age   int
	Other string
}

func (object *ObjectA) BeforeCreate(tx *gorm.DB) (err error) {
	object.ID, _ = uuid.NewUUID()
	return
}

// nolint:maintidx // Acceptable
func TestGormLike_Initialize_TriggersLikingCorrectly(t *testing.T) {
	t.Parallel()

	defaultQuery := func(db *gorm.DB) *gorm.DB { return db }

	tests := map[string]struct {
		filter   map[string]any
		existing []ObjectA
		options  []Option
		query    func(*gorm.DB) *gorm.DB

		expected []ObjectA
	}{
		"nothing": {
			expected: []ObjectA{},
			query:    defaultQuery,
		},

		// Check if everything still works
		"simple where query": {
			filter: map[string]any{
				"name": "jessica",
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 46}, {Name: "amy", Age: 35}},
			expected: []ObjectA{{Name: "jessica", Age: 46}},
		},
		"more complex where query": {
			filter: map[string]any{
				"name": "jessica",
				"age":  53,
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "jessica", Age: 20}},
			expected: []ObjectA{{Name: "jessica", Age: 53}},
		},
		"multi-value where query": {
			filter: map[string]any{
				"name": []string{"jessica", "amy"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}},
		},
		"more complex multi-value where query": {
			filter: map[string]any{
				"name": []string{"jessica", "amy"},
				"age":  []int{53, 20},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}},
		},

		// On to the 'real' tests
		"simple like query": {
			filter: map[string]any{
				"name": "%a%",
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}},
		},
		"more complex like query": {
			filter: map[string]any{
				"name": "%a%",
				"age":  20,
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "amy", Age: 20}},
		},
		"multi-value, all like queries": {
			filter: map[string]any{
				"name": []string{"%a%", "%o%"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
		},
		"more complex multi-value, all like queries": {
			filter: map[string]any{
				"name":  []string{"%a%", "%o%"},
				"other": []string{"%ooo", "aaa%"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aaaooo"}, {Name: "amy", Age: 20, Other: "aaaooo"}, {Name: "John", Age: 25, Other: "aaaooo"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "aaaooo"}, {Name: "amy", Age: 20, Other: "aaaooo"}, {Name: "John", Age: 25, Other: "aaaooo"}},
		},
		"multi-value, some like queries": {
			filter: map[string]any{
				"name": []string{"jessica", "%o%"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "John", Age: 25}},
		},
		"more complex multi-value, some like queries": {
			filter: map[string]any{
				"name":  []string{"jessica", "%o%"},
				"other": []string{"aa%", "bb"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}, {Name: "amy", Age: 20}, {Name: "John", Age: 25, Other: "bb"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}, {Name: "John", Age: 25, Other: "bb"}},
		},
		"explicitly disable liking in query": {
			filter: map[string]any{
				"name":  []string{"jessica", "%o%"},
				"other": []string{"aa%", "bb"},
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, false)
			},
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}, {Name: "amy", Age: 20}, {Name: "John", Age: 25, Other: "bb"}},
			expected: []ObjectA{},
		},

		// With custom character
		"simple like query with 🍌": {
			filter: map[string]any{
				"name": "🍌a🍌",
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}},
			options:  []Option{WithCharacter("🍌")},
		},
		"more complex like query with 🍓": {
			filter: map[string]any{
				"name": "🍓a🍓",
				"age":  20,
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "amy", Age: 20}},
			options:  []Option{WithCharacter("🍓")},
		},
		"multi-value, all like queries with 🍎": {
			filter: map[string]any{
				"name": []string{"🍎a🍎", "🍎o🍎"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			options:  []Option{WithCharacter("🍎")},
		},
		"more complex multi-value, all like queries with 🍎": {
			filter: map[string]any{
				"name":  []string{"🍎a🍎", "🍎o🍎"},
				"other": []string{"🍎ooo", "aaa🍎"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aaaooo"}, {Name: "amy", Age: 20, Other: "aaaooo"}, {Name: "John", Age: 25, Other: "aaaooo"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "aaaooo"}, {Name: "amy", Age: 20, Other: "aaaooo"}, {Name: "John", Age: 25, Other: "aaaooo"}},
			options:  []Option{WithCharacter("🍎")},
		},
		"multi-value, some like queries with 🍐": {
			filter: map[string]any{
				"name": []string{"jessica", "🍐o🍐"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}, {Name: "John", Age: 25}},
			options:  []Option{WithCharacter("🍐")},
		},
		"more complex multi-value, some like queries with 🍐": {
			filter: map[string]any{
				"name":  []string{"jessica", "🍐o🍐"},
				"other": []string{"aa🍐", "bc"},
			},
			query:    defaultQuery,
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}, {Name: "amy", Age: 20, Other: "bc"}, {Name: "John", Age: 25, Other: "bb"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}},
			options:  []Option{WithCharacter("🍐")},
		},

		// With existing query
		"simple like query with existing calls": {
			filter: map[string]any{
				"name": "%a%",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("other = ?", "goodbye")
			},
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "hello"}, {Name: "amy", Age: 20, Other: "goodbye"}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "amy", Age: 20, Other: "goodbye"}},
		},
		"more complex like query with existing calls": {
			filter: map[string]any{
				"name": "%a%",
				"age":  20,
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("other = ?", "def")
			},
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "abc"}, {Name: "amy", Age: 20, Other: "def"}, {Name: "John", Age: 25, Other: "ghi"}},
			expected: []ObjectA{{Name: "amy", Age: 20, Other: "def"}},
		},
		"multi-value, all like queries with existing calls": {
			filter: map[string]any{
				"name": []string{"%a%", "%o%", "%e%"},
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("other = ?", "no").Or("other = ?", "yes").Or("other = ?", "maybe")
			},
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "no"}, {Name: "amy", Age: 20, Other: "yes"}, {Name: "John", Age: 25, Other: "maybe"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "no"}, {Name: "amy", Age: 20, Other: "yes"}, {Name: "John", Age: 25, Other: "maybe"}},
		},
		"more complex multi-value, all like queries with existing calls": {
			filter: map[string]any{
				"other": []string{"aaa%", "%ooo"},
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name LIKE ?", "%a%").Or("name LIKE ?", "%o%")
			},
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aaaooo"}, {Name: "amy", Age: 20, Other: "aaaooo"}, {Name: "John", Age: 25, Other: "aaaooo"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "aaaooo"}, {Name: "amy", Age: 20, Other: "aaaooo"}, {Name: "John", Age: 25, Other: "aaaooo"}},
		},
		"multi-value, some like queries with existing calls": {
			filter: map[string]any{
				"name": []string{"%essica"},
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ?", "jessica")
			},
			existing: []ObjectA{{Name: "jessica", Age: 53}, {Name: "amy", Age: 20}, {Name: "John", Age: 25}},
			expected: []ObjectA{{Name: "jessica", Age: 53}},
		},
		"more complex multi-value, some like queries with existing calls": {
			filter: map[string]any{
				"other": []string{"aa%", "bb"},
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("name = ?", "jessica").Or("name LIKE ?", "%o%")
			},
			existing: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}, {Name: "amy", Age: 20}, {Name: "John", Age: 25, Other: "bb"}},
			expected: []ObjectA{{Name: "jessica", Age: 53, Other: "aab"}, {Name: "John", Age: 25, Other: "bb"}},
		},
		"UUID leading to like query": {
			filter: map[string]any{
				"id": []string{"1*"},
			},
			options:  []Option{WithCharacter("*")},
			query:    defaultQuery,
			existing: []ObjectA{{ID: uuid.MustParse("11473f0d-0d09-4880-a343-31a3f5fc73b7"), Name: "jessica", Age: 53, Other: "aab"}, {ID: uuid.MustParse("cade0be6-abab-4494-a29c-d89c138eb16e"), Name: "amy", Age: 20}, {ID: uuid.MustParse("1a4b1209-61c0-4fd6-a8b7-128930e7c426"), Name: "John", Age: 25, Other: "bb"}},
			expected: []ObjectA{{ID: uuid.MustParse("11473f0d-0d09-4880-a343-31a3f5fc73b7"), Name: "jessica", Age: 53, Other: "aab"}, {ID: uuid.MustParse("1a4b1209-61c0-4fd6-a8b7-128930e7c426"), Name: "John", Age: 25, Other: "bb"}},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectA{})
			plugin := New(testData.options...)

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			// Act
			err := db.Use(plugin)

			// Assert
			assert.NoError(t, err)

			var actual []ObjectA
			err = testData.query(db).Where(testData.filter).Find(&actual).Error
			assert.NoError(t, err)

			if assert.Len(t, actual, len(testData.expected)) {
				for i := range testData.expected {
					assert.Equal(t, testData.expected[i].Name, actual[i].Name)
					assert.Equal(t, testData.expected[i].Age, actual[i].Age)
					assert.Equal(t, testData.expected[i].Other, actual[i].Other)
				}
			}
		})
	}
}

func TestGormLike_Initialize_TriggersLikingCorrectlyWithConditionalTag(t *testing.T) {
	t.Parallel()

	type ObjectB struct {
		Name  string `gormlike:"true"`
		Other string
	}

	tests := map[string]struct {
		filter   map[string]any
		existing []ObjectB
		expected []ObjectB
	}{
		"simple filter on allowed fields": {
			filter: map[string]any{
				"name": "jes%",
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "amy", Other: "def"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}},
		},
		"simple filter on disallowed fields": {
			filter: map[string]any{
				"other": "%b%",
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"multi-filter on allowed fields": {
			filter: map[string]any{
				"name": []string{"jes%", "%my"},
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "amy", Other: "def"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "amy", Other: "def"}},
		},
		"multi-filter on disallowed fields": {
			filter: map[string]any{
				"other": []string{"%b%", "%c%"},
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectB{})
			plugin := New(TaggedOnly())

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			// Act
			err := db.Use(plugin)

			// Assert
			assert.NoError(t, err)

			var actual []ObjectB
			err = db.Where(testData.filter).Find(&actual).Error
			assert.NoError(t, err)

			assert.Equal(t, testData.expected, actual)
		})
	}
}

func TestGormLike_Initialize_AlwaysIgnoresFieldsWithGormLikeFalse(t *testing.T) {
	t.Parallel()

	type ObjectB struct {
		Name  string
		Other string `gormlike:"false"`
	}

	tests := map[string]struct {
		filter   map[string]any
		existing []ObjectB
		expected []ObjectB
	}{
		"Normal filter works on never field": {
			filter: map[string]any{
				"other": "abc",
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
		},
		"simple filter on disallowed fields": {
			filter: map[string]any{
				"other": "%b%",
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"multi-filter on disallowed fields": {
			filter: map[string]any{
				"other": []string{"%b%", "%c%"},
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}, {Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectB{})
			plugin := New(TaggedOnly())

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			// Act
			err := db.Use(plugin)

			// Assert
			assert.NoError(t, err)

			var actual []ObjectB
			err = db.Where(testData.filter).Find(&actual).Error
			assert.NoError(t, err)

			assert.Equal(t, testData.expected, actual)
		})
	}
}

func TestGormLike_Initialize_TriggersLikingCorrectlyWithSetting(t *testing.T) {
	t.Parallel()

	type ObjectB struct {
		Name  string
		Other string
	}

	tests := map[string]struct {
		filter   map[string]any
		query    func(*gorm.DB) *gorm.DB
		existing []ObjectB
		expected []ObjectB
	}{
		"like with query set to true": {
			filter: map[string]any{
				"name": "jes%",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, true)
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{{Name: "jessica", Other: "abc"}},
		},
		"like with query set to false": {
			filter: map[string]any{
				"name": "jes%",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, false)
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"like with query set to random value": {
			filter: map[string]any{
				"name": "jes%",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, "yes")
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
		"like with query unset": {
			filter: map[string]any{
				"name": "jes%",
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectB{})
			plugin := New(SettingOnly())

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			db = testData.query(db)

			// Act
			err := db.Use(plugin)

			// Assert
			assert.NoError(t, err)

			var actual []ObjectB
			err = db.Where(testData.filter).Find(&actual).Error
			assert.NoError(t, err)

			assert.Equal(t, testData.expected, actual)
		})
	}
}

func TestGormLike_Initialize_ProcessUnknownFields(t *testing.T) {
	t.Parallel()

	type ObjectB struct {
		Name  string
		Other string
	}

	tests := map[string]struct {
		filter   map[string]any
		query    func(*gorm.DB) *gorm.DB
		existing []ObjectB
		expected []ObjectB
	}{
		"like with unknown field": {
			filter: map[string]any{
				"name":          "jes%",
				"unknown_field": false,
			},
			query: func(db *gorm.DB) *gorm.DB {
				return db.Set(tagName, true)
			},
			existing: []ObjectB{{Name: "jessica", Other: "abc"}},
			expected: []ObjectB{},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			db := gormtestutil.NewMemoryDatabase(t, gormtestutil.WithName(t.Name()))
			_ = db.AutoMigrate(&ObjectB{})
			plugin := New(SettingOnly())

			if err := db.CreateInBatches(testData.existing, 10).Error; err != nil {
				t.Error(err)
				t.FailNow()
			}

			db = testData.query(db)

			// Act
			err := db.Use(plugin)

			// Assert
			assert.NoError(t, err)

			var actual []ObjectB
			err = db.Where(testData.filter).Find(&actual).Error
			assert.Equal(t, "no such column: unknown_field", err.Error())
			assert.Nil(t, actual)
		})
	}
}
