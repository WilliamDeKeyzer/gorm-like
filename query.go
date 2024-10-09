package gormlike

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const tagName = "gormlike"

//nolint:gocognit,cyclop // Acceptable
func (d *gormLike) queryCallback(db *gorm.DB) {
	fmt.Printf("Starting queryCallback \n")
	// If we only want to like queries that are explicitly set to true, we back out early if anything's amiss
	settingValue, settingOk := db.Get(tagName)
	if d.conditionalSetting && !settingOk {
		return
	}

	if settingOk {
		if boolValue, _ := settingValue.(bool); !boolValue {
			return
		}
	}

	fmt.Printf("Starting queryCallback 2\n")
	exp, settingOk := db.Statement.Clauses["WHERE"].Expression.(clause.Where)
	if !settingOk {
		fmt.Printf("early stop at where\n")
		return
	}

	fmt.Printf("exp.Exprs = %v", exp.Exprs)
	for index, cond := range exp.Exprs {
		switch cond := cond.(type) {
		case clause.Eq:
			fmt.Printf("clause eq\n")
			columnName, columnOk := cond.Column.(string)
			if !columnOk {
				continue
			}
			fmt.Printf("column name = %s\n", columnName)

			// Get the `gormlike` value
			var tagValue string
			dbField, ok := db.Statement.Schema.FieldsByDBName[columnName]
			if ok {
				tagValue = dbField.Tag.Get(tagName)
			}

			// If the user has explicitly set this to false, ignore this field
			if tagValue == "false" {
				continue
			}

			// If tags are required and the tag is not true, ignore this field
			if d.conditionalTag && tagValue != "true" {
				continue
			}

			value, columnOk := cond.Value.(string)
			if !columnOk {
				continue
			}

			// If there are no % AND there aren't ony replaceable characters, just skip it because it's a normal query
			if !strings.Contains(value, "%") && !(d.replaceCharacter != "" && strings.Contains(value, d.replaceCharacter)) {
				continue
			}

			// UUID has LIKE implementation
			var condition string
			// if isLikeableField(dbField.DataType, dbField.FieldType) {
			if dbField.FieldType.String() != "uuid.UUID" {
				fmt.Printf("likeable field '%s' with type '%s'\n", dbField.Name, dbField.FieldType)
				condition = fmt.Sprintf("%s LIKE ?", cond.Column)
			} else {
				fmt.Printf("not likeable field '%s' with type '%s'\n", dbField.Name, dbField.FieldType)
				condition = fmt.Sprintf("CAST(%s as varchar) LIKE ?", cond.Column)
			}
			fmt.Printf("condition = %s\n", condition)

			if d.replaceCharacter != "" {
				value = strings.ReplaceAll(value, d.replaceCharacter, "%")
			}

			exp.Exprs[index] = db.Session(&gorm.Session{NewDB: true}).Where(condition, value).Statement.Clauses["WHERE"].Expression
		case clause.IN:
			fmt.Printf("clause ins\n")
			columnName, columnOk := cond.Column.(string)
			if !columnOk {
				continue
			}
			fmt.Printf("column name = %s\n", columnName)

			// Get the `gormlike` value
			var tagValue string
			dbField, ok := db.Statement.Schema.FieldsByDBName[columnName]
			if ok {
				tagValue = dbField.Tag.Get(tagName)
			}

			// If the user has explicitly set this to false, ignore this field
			if tagValue == "false" {
				continue
			}

			// If tags are required and the tag is not true, ignore this field
			if d.conditionalTag && tagValue != "true" {
				continue
			}

			var likeCounter int
			var useOr bool

			query := db.Session(&gorm.Session{NewDB: true})

			for _, value := range cond.Values {
				value, ok := value.(string)
				if !ok {
					continue
				}
				condition := fmt.Sprintf("%s = ?", cond.Column)

				// If there are no % AND there aren't ony replaceable characters, just skip it because it's a normal query
				if (strings.Contains(value, "%") && d.replaceCharacter == "") || (d.replaceCharacter != "" && strings.Contains(value, d.replaceCharacter)) {

					// UUID has LIKE implementation
					if dbField.FieldType.String() != "uuid.UUID" {
						// if isLikeableField(dbField.DataType, dbField.FieldType) {
						fmt.Printf("likeable field '%s' with type '%s'\n", dbField.Name, dbField.FieldType)
						condition = fmt.Sprintf("%s LIKE ?", cond.Column)
					} else {
						fmt.Printf("not likeable field '%s' with type '%s'\n", dbField.Name, dbField.FieldType)
						condition = fmt.Sprintf("CAST(%s as varchar) LIKE ?", cond.Column)
					}
					fmt.Printf("condition = %s\n", condition)

					if d.replaceCharacter != "" {
						value = strings.ReplaceAll(value, d.replaceCharacter, "%")
					}

					likeCounter++
				}

				if useOr {
					query = query.Or(condition, value)

					continue
				}

				query = query.Where(condition, value)
				useOr = true
			}

			// Don't alter the query if it isn't necessary
			if likeCounter == 0 {
				continue
			}

			fmt.Printf("Replacing with where = %v\n", query)
			exp.Exprs[index] = db.Session(&gorm.Session{NewDB: true}).Where(query).Statement.Clauses["WHERE"].Expression
		}
	}
}

// func isLikeableField(dataType schema.DataType, fieldType reflect.Type) bool {
// 	fmt.Printf("isLikeableField with fieldtype '%s' and dataType '%v'\n", fieldType, dataType)
// 	if fieldType.String() == "uuid.UUID" {
// 		return false
// 	}
// 	return dataType == schema.String || dataType == schema.Bool || dataType == schema.Int || dataType == schema.Uint || dataType == schema.Float || dataType == schema.Time || dataType == schema.Bytes
// }
