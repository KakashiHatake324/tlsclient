package tlsclient

import "sync"

var SitePrints = [][2]string{
	{"footsextra", "67add1166b020ae61b8f5fc96813c04c2aa589960796865572a3c7e737613dfd"},
	{"footsextra", "6d99fb265eb1c5b3744765fcbc648f3cd8e1bffafdc4c2f99b9d47cf7ff1c24f"},
	{"snipesextra", "154c433c491929c5ef686e838e323664a00e6a0d822ccc958fb4dab03e49a08f"},
	{"walmartextra", "b676ffa3179e8812093a1b5eafee876ae7a6aaf231078dad1bfb21cd2893764a"},
	{"walmartextra", "cbb522d7b7f127ad6a0113865bdf1cd4102e7d0759af635a7cf4720dc963c53b"},
	{"finishlineextra", "8cc34e11c167045824ade61c4907a6440edb2c4398e99c112a859d661f8e2bc7"},
	{"fastlyextra", "1dd095449ffb3cff14b2224d596b83fd42f2b47683553c797d11150c918691bd"},
	{"footlocker", "a12a634e3ca34d1aeda7b7fd63368bf38489137d24aa1fee8b9cfc2e474be162"},
	{"footaction", "79a26bb888e285d2b5eec5e6cf84c18dfe5cac9fd496df77b4081a5eec2ec0b2"},
	{"eastbay", "b2262e09f2d8c0e86608587105b551c410f945c5da3e750189b4ca3a62d03b6e"},
	{"champssports", "d81b0c1424f07370c3f4efa21460f054b5e1d5abbd27c025f07f96eb827b34a1"},
	{"kidsfootlocker", "4095b37929dd175ae9621d086e1b782c8934da092efa98847a2f10b9c69bd454"},
	{"footlockerca", "1ba2f8b6d2dbc338b42a30e5d1c0b9c1774863875d7ea7108409f976b890eccd"},
	{"snipesusa", "fe32ad15340cde43ad08bbd8886536ea0e01608842a53374bf179b046ebc1a4e"},
	{"snipesusaextra", "7fa4ff68ec04a99d7528d5085f94907f4d1dd1c5381bacdc832ed5c960214676"},
	{"walmart", "d156c5a2911d7a3919e8025669592d3224ccc02a87db678ea026f51dab9d5d02"},
	{"finishline", "f5c18ce512295d3ba7e39bbae1bbf8afc8ab873931982f0470b4f84770ffc0a2"},
	{"jdsports", "378cc09c49c7bca44809359a8e997c9094d343ca16ff854f81d2a7e1d7acb248"},
	{"yeezysupply", "db6328e23bc59b41ce9247a1e382b5a12f7266ddac6a7a2b9dc9cafe3af26d89"},
	{"ladyfootlocker", "dd843ea7c8d376593a25748a72f648cd1f444ba2c2cec4a665346af5b99311f8"},
	{"cachenodes", "6071491e11e842fd18d7e371919b1b3425ee11f0c6fbf93dbc945624babb9aa5"},
	{"fastly", "e16ebac3d26248739d712566abd9763379d1b4b5d219f799355795ad3d5611d9"},
	{"queueit", "e7ddf7ded151b3e3520b342aae72b6f9b8c42ef72ac851f032fa9895b4e93fd3"},
	{"queueit", "f55f9ffcb83c73453261601c7e044db15a0f034b93c05830f28635ef889cf670"},
	{"queueit", "28689b30e4c306aab53b027b29e36ad6dd1dcf4b953994482ca84bdc1ecac996"},
	{"hibbett", "3bf8655fa3cc32404a842539d9ac39ddded41a5a2f3e082ab63d50fbde11ef1a"},
	{"hibpay", "a8008df37ff4c827a6dad7e0cca16476ebb4d4fb0322c3d0451741a6ede98ef0"},
	{"hawk", "58acca84e5da1ba2926a43ca32ef2e619ff7a2ef0bc0ea2ef256566351d0e4ad"},
	{"hawk", "87dcd4dc74640a322cd205552506d1be64f12596258096544986b4850bc72706"},
	{"hawk", "2072cbd9014a1ccb72992bd95a0c84b7ffdc53b76a16ab417d7fbfbec44cc479"},
	{"hawk", "4348a0e9444c78cb265e058d5e8944b4d84f9662bd26db257f8934a443c70161"},
}

var loadedCerts = make(map[string]string)
var certMutex sync.RWMutex
