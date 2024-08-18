package main

import "github.com/vmihailenco/msgpack/v5"

type BidRequest struct {
	Bid    uint64 `json:"bid"`
	Bidder string `json:"bidder"`
}

func (req BidRequest) MarshalBinary() ([]byte, error) {
	type tmp struct {
		Bid      uint64
		BidderID string
	}
	return msgpack.Marshal(tmp{Bid: req.Bid, BidderID: req.Bidder})
}

func (req *BidRequest) UnmarshalBinary(data []byte) error {
	type tmp struct {
		Bid      uint64
		BidderID string
	}
	var bfr tmp
	if err := msgpack.Unmarshal(data, &bfr); err != nil {
		return err
	}
	req.Bid = bfr.Bid
	req.Bidder = bfr.BidderID
	return nil
}

type BidNotification struct {
	ItemID string
	Info   BidRequest
}
