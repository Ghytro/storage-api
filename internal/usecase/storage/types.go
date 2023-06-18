package storage

import "errors"

// storage schema definition types

type StorageSchemaReq []StorageSchemaReqItem

func (req StorageSchemaReq) Validate() error {
	for _, r := range req {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type StorageSchemaReqItem struct {
	IsAvailable bool                      `json:"is_available"`
	Products    []StorageSchemaReqProduct `json:"products"`
}

func (i StorageSchemaReqItem) Validate() error {
	for _, p := range i.Products {
		if err := p.Validate(); err != nil {
			return err
		}
	}
	// check if all the vendors are unique
	{
		m := map[string]struct{}{}
		for _, p := range i.Products {
			m[p.Vendor] = struct{}{}
		}
		if len(m) != len(i.Products) {
			return errors.New("not all the vendors are unique")
		}
	}
	return nil
}

type StorageSchemaReqProduct struct {
	Vendor string `json:"vendor"`
	Name   string `json:"name"`
	Size   string `json:"size"`
	Amount uint   `json:"amount"`
}

func (p StorageSchemaReqProduct) Validate() error {
	if len(p.Vendor) > 20 {
		return errors.New("vendor length too big")
	}
	if p.Amount == 0 {
		return errors.New("added product amount cannot be zero")
	}
	return nil
}

type StorageSchemaResp []StorageSchemaRespItem

type StorageSchemaRespItem struct {
	ID          uint                       `json:"id"`
	IsAvailable bool                       `json:"is_available"`
	Products    []StorageSchemaRespProduct `json:"products"`
}

type StorageSchemaRespProduct struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Vendor string `json:"vendor"`
	Size   string `json:"size"`
}
