<template>
	<div style="overflow:auto" class="search">
		<div class="menu">
			<a v-for="item in menuItems" v-bind:key="item.Name" :href="item.Url">{{ item.Name }}</a>
		</div>

		<div class="main">
			<h2>Search</h2>
			<label>Search for books</label>
			<input size="50" v-model="searchTerm">

			<div class="result">
				<table style="width:100%" border="2"  v-for="searchResult in searchResults" v-bind:key="searchResult.Name" >
					<tr>
						<td>Name</td>
						<td>{{ searchResult.Name }} </td>
					</tr>
					<tr>
						<td>Author</td>
						<td>{{ searchResult.Author }} </td>
					</tr>
				</table>
			</div>
		</div>



		<div class="right">
			<h2>About</h2>
			<p>Lend books to read.</p>
		</div>

	</div>
</template>

<script src="https://cdn.jsdelivr.net/npm/axios@0.12.0/dist/axios.min.js"></script>
<script>
  import axios from 'axios';
  export default {
    name: 'Search',
    props: {
      msg: String
    },
    data: function () {
      return {
        searchResultObject: {
          Name:'',
					Id: '',
					Author: '',
				},
        searchTerm: '',
				erroMessage: '',
				searchResults: [],
        menuItems: [
          {Name: "Home", Url: "#"},
          {Name: "New Arrivals", Url: "#"},
          {Name: "Sponsered Books", Url: "#"},
          {Name: "Sponsered Books 1", Url: "#"},
        ]
      }
		},
    created: function () {
        this.getBook()
    },
    methods: {
      getBook: function () {
        var vm = this
        axios.get('http://localhost:9090/library/v1/book')
            .then(function (response) {
              vm.searchResults =response.data
            })
            .catch(function (error) {
              vm.erroMessage = 'Error! Could not reach the API. ' + error
            })
      }
    }

  }
</script>


<style scoped>
	* {
		box-sizing: border-box;
	}
	.menu {
		float:left;
		width:20%;
		text-align:center;
	}
	.menu a {
		background-color:#e5e5e5;
		padding:8px;
		margin-top:7px;
		display:block;
		width:100%;
		color:black;
	}
	.main {
		float:left;
		width:60%;
		padding:0 20px;
	}

	.main result {
		float:left;
		width:60%;
		padding:0 20px;
	}
	.right {
		background-color:#e5e5e5;
		float:left;
		width:20%;
		padding:15px;
		margin-top:7px;
		text-align:center;
	}

	table, th, td {
		border: 1px transparent;
		width:50%;
		text-align: left;
	}
	@media only screen and (max-width:620px) {
		/* For mobile phones: */
		.menu, .main, .right {
			width:100%;
		}
	}
</style>
