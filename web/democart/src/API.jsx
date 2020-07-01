import axios from "axios";

// to connect to a development server, uncomment
import packageJSON from "../package.json";
const SERVER_URL = packageJSON.server;

const AUTH_URL = {
  SIGNUP: `${SERVER_URL}/auth/signup`,
  LOGIN:  `${SERVER_URL}/auth/login`,
  LOGOUT: `${SERVER_URL}/auth/logout`,
};

const API_URL = {
  HEALTH:   `${SERVER_URL}`,
  USER:     `${SERVER_URL}/api`,
  ADDRESS:  `${SERVER_URL}/api/address`,
  ITEM:     `${SERVER_URL}/api/item`,
  CART:     `${SERVER_URL}/api/cart`,
  ORDER:    `${SERVER_URL}/api/order`,
};

// TODO(sam): look into using fetch instead of axios?
const baseAPI = {
  makeRequest(method, url, accessToken, params, data, success, failure) {
    let headers = {};
    if (accessToken) {
      headers = {'Authorization': 'Bearer ' + accessToken};
    }

    // returns the promise so each api response can take advantage of 'finally'
    return axios({
      url: url,
      method: method,
      withCredentials: true,
      params: params || {},
      data: data || {},
      headers: headers,
    })
    .then(response => {
      if (!success) return response;
      // https://github.com/axios/axios#response-schema
      return success(response.data);
    })
    .catch(err => {
      // TODO(sam): this would be the place for a refresh token exchange
      // attempt and then a redo of the request
      if (!failure) {
        let e = err;
        if (err && err.response && err.response.data) {
          e = err.response.data;
        }
        console.error("Error: ", e);
        return e;
      }
      return failure(err.response);
    });
  },

  makeGetRequest(url, accessToken, params, data, success, failure) {
    return this.makeRequest('get', url, accessToken, params, data, success,
      failure);
  },

  makePostRequest(url, accessToken, params, data, success, failure) {
    return this.makeRequest('post', url, accessToken, params, data, success,
      failure);
  },

  makeDeleteRequest(url, accessToken, params, data, success, failure) {
    return this.makeRequest('delete', url, accessToken, params, data, success,
      failure);
  },
}

const badAccessTokenNotification = () => {
  // TODO(sam): return some kind of notification that the resource needs a
  // valid access token, but the provided one is bad.
  return null;
}

export default {
  health(success) {
    return baseAPI.makeGetRequest(API_URL.HEALTH, null, null, null,
      success, null);
  },

  authCodeExchange(redirectURI, code, state, success, failure) {
    return baseAPI.makeGetRequest(redirectURI, null,
      {code: code, state: state}, null, success, failure);
  },

  getLogin() {
    return AUTH_URL.LOGIN;
  },

  getSignup() {
    return AUTH_URL.SIGNUP;
  },

  logout(accessToken,) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makeGetRequest(AUTH_URL.LOGOUT, accessToken, null, null,
      null, null);
  },

  getUser(accessToken, success, failure) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makeGetRequest(API_URL.USER, accessToken, null, null,
      success, failure);
  },

  makeAddress(accessToken, address, success) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makePostRequest(API_URL.ADDRESS, accessToken, null,
      address, success, null);
  },

  allItems(success) {
    return baseAPI.makeGetRequest(API_URL.ITEM, null, null, null,
      success, null);
  },

  updateItem(accessToken, item, success) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makePostRequest(API_URL.ITEM + "/" + item.id,
      accessToken, null, item, success, null);
  },

  makeItem(accessToken, item, success) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makePostRequest(API_URL.ITEM, accessToken, null,
      item, success, null);
  },

  cart(accessToken, success) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makeGetRequest(API_URL.CART, accessToken, null, null,
      success, null);
  },

  updateCart(accessToken, cartItem, success, failure) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makePostRequest(API_URL.CART + "/" + cartItem.item_id,
      accessToken, null, cartItem, success, failure);
  },

  addToCart(accessToken, cartItem, success, failure) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makePostRequest(API_URL.CART, accessToken, null,
      cartItem, success, failure);
  },

  allOrders(accessToken, success) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makeGetRequest(API_URL.ORDER, accessToken, null, null,
      success, null);
  },

  orderCart(accessToken, items, success, failure) {
    if (!accessToken) { return badAccessTokenNotification(); }
    return baseAPI.makePostRequest(API_URL.ORDER, accessToken, null,
      {ordered_items: items}, success, failure);
  },
}
