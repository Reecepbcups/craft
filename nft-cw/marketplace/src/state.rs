use crate::msg::{ContractInformationResponse};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

use cosmwasm_std::{Addr, StdResult, Storage, Uint128};
// use cosmwasm_std::Coin;
use cw_storage_plus::{Index, IndexList, IndexedMap, Item, Map, MultiIndex};

pub static _CONFIGKEY: &[u8] = b"config";

#[derive(Serialize, Deserialize, Clone, PartialEq, JsonSchema, Debug)]
pub struct Offering {
    pub token_id: String,
    pub list_denom: String,
    pub contract_addr: Addr,
    pub token_uri: String,
    pub seller: Addr,
    pub list_price: Uint128,
}

#[derive(Serialize, Deserialize, Clone, PartialEq, JsonSchema, Debug)]
pub struct Volume {
    pub collection_volume: Uint128,
    pub num_traded: Uint128,
}

/// OFFERINGS is a map which maps the offering_id to an offering. Offering_id is derived from OFFERINGS_COUNT.
pub const OFFERINGS: Map<&str, Offering> = Map::new("offerings");

pub const COLLECTION_VOLUME: Map<&str, Volume> = Map::new("collection_volume");

// we only keep storage of last 20, pop()'ed off when a new NFT is bought from any collection
pub const RECENTLY_SOLD: Item<Vec<Offering>> = Item::new("recently_sold");

pub const OFFERINGS_COUNT: Item<u64> = Item::new("num_offerings");

pub const CONTRACT_INFORMATION: Item<ContractInformationResponse> = Item::new("marketplace_information");

// Offerings helpers
pub fn num_offerings(storage: &dyn Storage) -> StdResult<u64> {
    Ok(OFFERINGS_COUNT.may_load(storage)?.unwrap_or_default())
}

pub fn increment_offerings(storage: &mut dyn Storage) -> StdResult<u64> {
    let val = num_offerings(storage)? + 1;
    OFFERINGS_COUNT.save(storage, &val)?;
    Ok(val)
}

pub struct OfferingIndexes<'a> {
    pub seller: MultiIndex<'a, Addr, Offering, (Addr, Vec<u8>)>,
    pub contract: MultiIndex<'a, Addr, Offering, (Addr, Vec<u8>)>,
}

impl<'a> IndexList<Offering> for OfferingIndexes<'a> {
    fn get_indexes(&'_ self) -> Box<dyn Iterator<Item = &'_ dyn Index<Offering>> + '_> {
        let v: Vec<&dyn Index<Offering>> = vec![&self.seller, &self.contract];
        Box::new(v.into_iter())
    }
}

pub fn offerings<'a>() -> IndexedMap<'a, &'a str, Offering, OfferingIndexes<'a>> {
    let indexes = OfferingIndexes {
        seller: MultiIndex::new(
            |o: &Offering| o.seller.clone(),
            "offerings",
            "offerings__seller",
        ),

        contract: MultiIndex::new(
            |o: &Offering| o.contract_addr.clone(),
            "offerings",
            "offerings__contract",
        ),
    };
    IndexedMap::new("offerings", indexes)
}
