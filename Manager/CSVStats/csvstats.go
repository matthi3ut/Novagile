package CSVStats

import (
	"encoding/csv"
	"fmt"
	pcsv "github.com/lpuig/Novagile/Manager/ProcessCSV"
	"io"
)

type IndexDesc struct {
	name string
	cols []string
}

func NewIndexDesc(name string, cols ...string) IndexDesc {
	return IndexDesc{name: name, cols: cols}
}

type CSVStats struct {
	data       *pcsv.CSVStats
	indexDescs []IndexDesc
	indexes    map[string]*index
}

func NewCSVStats(indexes ...IndexDesc) *CSVStats {
	c := &CSVStats{}
	c.data = pcsv.NewCSVStats()
	c.indexDescs = indexes
	c.indexes = map[string]*index{}
	return c
}

func (c *CSVStats) AddHeader(record []string) error {
	c.data.AddHeader(record)
	h := c.data.GetHeader()
	for _, id := range c.indexDescs {
		rs, err := h.NewRecordSelector(id.cols...)
		if err != nil {
			return fmt.Errorf("CSVStats Unable to create index '%s' : %s", id.name, err.Error())
		}
		c.indexes[id.name] = newIndex(rs)
	}
	return nil
}

// Add adds the given record, updating indexes
func (c *CSVStats) Add(record []string) {
	num := c.data.Add(record)
	for _, v := range c.indexes {
		v.Add(record, num)
	}
}

// GetKeys returns all registered keys for given idxname
func (c *CSVStats) GetKeys(idxname string) []string {
	if i, found := c.indexes[idxname]; found {
		return i.index.Keys()
	}
	return nil
}

func (c *CSVStats) GetRecords(idxname, key string) [][]string {
	i, found := c.indexes[idxname]
	if !found {
		return nil
	}
	r, found := i.index[key]
	if !found {
		return nil
	}
	res := [][]string{}
	for _, pos := range r {
		res = append(res, c.data.Get(pos))
	}
	return res
}

// AddCSVDataFrom populates the CSVStats with Data from given reader (CSV formated data) (Header and Indexes are populated)
func (c *CSVStats) AddCSVDataFrom(r io.Reader) error {
	csvr := csv.NewReader(r)
	csvr.Comma = ';'
	csvr.Comment = '#'

	var numline int = 0
	for {
		record, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if numline == 0 {
			err = c.AddHeader(record)
			if err != nil {
				return err
			}
		} else {
			c.Add(record)
		}

		numline++
	}
	return nil
}
