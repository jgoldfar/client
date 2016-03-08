// @flow
import React, {Component} from 'react'
import {TextField} from 'material-ui'
import {globalStyles, globalColorsDZ2} from '../styles/style-guide'
import {styles as TextStyles, specialStyles} from './text'
import materialTheme from '../styles/material-theme.desktop'

import type {Props} from './input'

export default class Input extends Component {
  props: Props;
  state: {value: ?string, focused: boolean};
  _textField: any;

  constructor (props: Props) {
    super(props)

    this.state = {
      value: props.value,
      focused: false
    }
  }
  getChildContext (): Object {
    return {
      muiTheme: materialTheme
    }
  }

  getValue (): ?string {
    return this.state.value
  }

  clearValue () {
    this.setState({value: null})
  }

  onChange (event: {target: {value: ?string}}) {
    this.setState({value: event.target.value})
  }

  blur () {
    this._textField && this._textField.blur()
  }

  render () {
    const style = this.props.small ? styles.containerSmall : styles.container
    const textStyle = this.props.small ? styles.inputSmall : styles.input

    // HACK: We can't reset the text area style, so we need to counteract it by moving the wrapper up
    const multiLineStyleFix = {
      height: 'auto', position: 'relative',
      // Other HACK: having a floating label affects position, but only in multiline
      bottom: (this.props.floatingLabelText ? 30 : 5),
      marginTop: 6
    }
    const inputStyle = this.props.multiLine ? multiLineStyleFix : {height: 'auto'}
    return (
      <div style={{...style, ...this.props.style}} onClick={() => { this._textField && this._textField.focus() }}>
        <TextField
          ref={textField => (this._textField = textField)}
          fullWidth
          inputStyle={{...inputStyle, textAlign: 'center'}}
          underlineStyle={{borderColor: globalColorsDZ2.black10, bottom: 'auto'}}
          errorStyle={{...styles.errorStyle, ...this.props.errorStyle}}
          style={{...textStyle, ...globalStyles.flexBoxColumn}}
          autoFocus={this.props.autoFocus}
          errorText={this.props.errorText}
          floatingLabelText={this.props.small ? undefined : this.props.floatingLabelText}
          floatingLabelStyle={{...styles.floatingLabelStyle, ...(this.state.value || this.state.focused ? {color: globalColorsDZ2.blue, transform: 'perspective(1px) scale(0.64) translate3d(2px, -28px, 0)', transformOrigin: 'center top'} : {transform: 'scale(1) translate3d(0, 0, 0)'})}}
          onFocus={() => this.setState({focused: true})}
          onBlur={() => this.setState({focused: false})}
          hintText={this.props.hintText}
          hintStyle={{...styles.hintStyle, ...(this.props.multiLine ? {textAlign: 'center'} : {top: 3, bottom: 'auto'})}}
          multiLine={this.props.multiLine}
          onChange={event => {
            this.onChange(event)
            this.props.onChange && this.props.onChange(event)
          }}
          onEnterKeyDown={this.props.onEnterKeyDown}
          underlineFocusStyle={styles.underlineFocusStyle}
          rows={this.props.rows}
          rowsMax={this.props.rowsMax}
          type={this.props.type}
          value={this.state.value}
          />
      </div>
    )
  }
}

Input.childContextTypes = {
  muiTheme: React.PropTypes.object
}

export const styles = {
  container: {
    marginBottom: 8
  },
  containerSmall: {
    margin: 0,
    marginTop: 2
  },
  input: {
    ...specialStyles.textInput,
    height: 80
  },
  inputSmall: {
    ...TextStyles.textBody,
    ...TextStyles.textSmallMixin,
    height: 40,
    lineHeight: '11px'
  },
  underlineFocusStyle: {
    borderColor: globalColorsDZ2.blue,
    borderBottom: 'solid 1px',
    transition: ''
  },
  errorStyle: {
    ...globalStyles.DZ2.fontRegular,
    color: globalColorsDZ2.red,
    alignSelf: 'center',
    fontSize: 14,
    lineHeight: '17px',
    position: 'initial',
    marginTop: 4,
    paddingTop: 4
  },
  hintStyle: {
    ...globalStyles.DZ2.fontRegular
  },
  floatingLabelStyle: {
    ...globalStyles.DZ2.fontRegular,
    color: globalColorsDZ2.black10,
    alignSelf: 'center',
    position: 'inherit',
    top: 34
  }
}
