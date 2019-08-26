# Golang package for reading and writing dbf files

[![Build Status](https://travis-ci.org/kcasctiv/dbf3.svg?branch=master)](https://travis-ci.org/kcasctiv/dbf3)
[![GoDoc](https://godoc.org/github.com/kcasctiv/dbf3?status.svg)](https://godoc.org/github.com/kcasctiv/dbf3)

README is currently under construction

## Examples

```Go
// Open (from file)
file, err := dbf3.OpenFile("filename.dbf")

// Open (from reader)
file, err := dbf3.Open(reader)

// Create new
file := dbf3.New(dbf3.WithLang(langDriver))

// Change language driver
file.SetLang(newDriver)

// Get values
fields := file.Fields()
for idx := 0; idx < file.Rows(); idx++ {
    for _, field := range fields {
        value, _ := file.Get(idx, field.Name())
        fmt.Println(value)
    }
}

// Set value
err := file.Set(idx, "field_name", "value")

// Add row
idx, err := file.NewRow()

// Delete row
err := file.DelRow(idx)

// Check row is deleted
deleted, err := file.Deleted(idx)

// Add field
err := file.AddField("field_name", dbf3.Character, length, decimals)

// Delete field
err := file.DelField("field_name")

// Save (into file)
err := file.SaveFile("filename.dbf")

// Save (into writer)
err := file.Save(writer)
```

## Next steps (random order)

* GoDoc
* Tests
* Examples
* Benchmarks?
* Go badges
* Check values when set
* Add zeros to the floats end (for padding by scale)
* Add more checks for `AddField`
* Complete README
* Add missed charsets
* Go modules
