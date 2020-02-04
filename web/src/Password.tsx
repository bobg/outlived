import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'
import { Label } from 'semantic-ui-react'

interface Props {
  buttonLabel?: string
  onClose: () => void
  onForgot?: () => void
  onSubmit: (pw: string) => void
  prompt?: string
  show: () => boolean
  title?: string
}

interface State {
  pw: string
}

export class PasswordDialog extends React.Component<Props, State> {
  private inp?: HTMLInputElement

  public state = { pw: '' }

  private handleSubmit = (ev: React.FormEvent<HTMLFormElement>) => {
    const { pw } = this.state
    if (!passwordValid(pw)) {
      return
    }
    this.props.onSubmit(this.state.pw)
    this.props.onClose()
    ev.preventDefault()
  }

  private handleForgot = (ev: React.FormEvent<HTMLButtonElement>) => {
    ev.preventDefault()
    if (!this.props.onForgot) {
      return
    }
    this.props.onForgot()
    this.props.onClose()
  }

  private handleChange = (ev: React.FormEvent<HTMLInputElement>) => {
    this.setState({ pw: ev.currentTarget.value })
  }

  public render = () => {
    const { buttonLabel, prompt, title } = this.props
    return (
      <Modal
        onEntered={() => {
          if (this.inp) {
            this.inp.focus()
          }
        }}
        onHide={this.props.onClose}
        show={this.props.show()}
        id='password-modal'
      >
        <Modal.Header closeButton>
          <ModalTitle>{title || 'Password'}</ModalTitle>
        </Modal.Header>
        <ModalBody>
          <form onSubmit={this.handleSubmit}>
            <Label>
              {prompt || 'Enter password'}{' '}
              <input
                type='password'
                id='password'
                ref={(inp: HTMLInputElement) => {
                  this.inp = inp
                }}
                value={this.state.pw}
                onChange={this.handleChange}
              />
            </Label>
            <br />
            <button type='submit' disabled={!passwordValid(this.state.pw)}>
              {buttonLabel || 'Submit'}
            </button>
            {this.props.onForgot && (
              <button onClick={this.handleForgot}>Forgot password</button>
            )}
          </form>
        </ModalBody>
      </Modal>
    )
  }
}

const passwordValid = (inp: string) => {
  return !!inp
}
