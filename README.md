# Democart

This is an exercise in building a modular and containerized Shopping Cart
service. It consists of 3 main components:

1. A backend REST API, written in Go.
2. An identity provider, written in Go.
3. A frontend marketplace, using ReactJS.

The backend REST API and the identity provider are both running on the same
host for simplicity, but they could easily be broken out. They are both located
in `go/src/democart`.  
The frontend marketplace is located in `web/democart`.


### Dependencies

- docker


### Building From Scratch and Running Locally

```sh
make new
```


### Running Locally

```sh
make up
```


### Start a Quick'n'Dirty Dev Server

```sh
make devup
```


### How to Use

- Go to [localhost:3000](http://localhost:3000) in a browser.
- "Sign Up" to create a new user
- "Make Dummy Address"
- "Make Dummy Item"
- See the item in the marketplace
- Add the item to your cart
- Add the item to your cart again
- Order the items in your cart
- Logout
- See the Dummy items in the marketplace


### TODOs

- More unit tests
- Better documentation
- More graceful error handling
- Stylize the frontend better
- Better configuration handling between psql secrets and server config file
