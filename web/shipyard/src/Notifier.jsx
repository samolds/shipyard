import React, {Component} from 'react'
import PropTypes from 'prop-types'


export default class Notifier extends Component {
  display = (n) => {
    if (n && n.response && n.response.data) {
      return <strong>{ n.response.data }</strong>
    }

    if (n && n.response && n.response.data) {
      return <strong>{ n.response.data }</strong>
    }

    if (n && n.response) {
      return <strong>{ n.response }</strong>
    }

    if (n && n.message) {
      return <strong>{ n.message }</strong>
    }

    if (n && n.data && n.data.error) {
      return <strong>{ n.data.error }</strong>
    }

    if (n && n.data) {
      return <strong>{ n.data }</strong>
    }

    return null;
  }

  render() {
    return (
      <div className="notifier-container">
        <p className="notification">{ this.display(this.props.error) }</p>
      </div>
    );
  }
}

Notifier.propTypes = {
  success: PropTypes.object,
  error: PropTypes.object,
}
