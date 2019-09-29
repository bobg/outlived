import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'

interface Props {
  buttonLabel?: string
  onClose: () => void
  onSubmit: (pw: string) => void
  prompt?: string
  show: () => boolean
  title?: string
}

interface State {
  pw: string
}

export class PasswordDialog extends React.Component<Props, State> {
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

  private handleChange = (ev: React.FormEvent<HTMLInputElement>) => {
    this.setState({ pw: ev.currentTarget.value })
  }

  public componentDidMount = () => {
    this.setState({ pw: '' })
  }

  public render = () => {
    const { buttonLabel, prompt, title } = this.props
    return (
      <Modal onHide={this.props.onClose} show={this.props.show()}>
        <Modal.Header closeButton>
          <ModalTitle>{title || 'Password'}</ModalTitle>
        </Modal.Header>
        <ModalBody>
          <form onSubmit={this.handleSubmit}>
            <label htmlFor='password'>
              {prompt || 'Enter password'}
              <input
                type='password'
                id='password'
                value={this.state.pw}
                onChange={this.handleChange}
              />
            </label>
            <button
              type='submit'
              disabled={!passwordValid(this.state.pw)}
            >
              {buttonLabel || 'Submit'}
            </button>
          </form>
        </ModalBody>
      </Modal>
    )
  }
}

const passwordValid = (inp: string) => {
  return !!inp
}