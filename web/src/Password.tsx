import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'

interface Props {
  buttonLabel: string
  // xxx add oncancel callback
  onSubmit: (pw: string) => void
  prompt?: string
  title?: string
}

interface State {
  pw: string
  show: boolean
}

export class PasswordDialog extends React.Component<Props, State> {
  public state = { pw: '', show: true }

  private handleSubmit = () => {
    const { pw } = this.state
    if (!passwordValid(pw)) {
      return
    }
    this.onSubmit(this.state.pw)
    this.setState({ show: false })
  }

  private handleChange = (ev: Event) => {
    this.setState({ pw: ev.target.value })
  }

  public componentDidMount = () => {
    this.setState({ pw: '', show: true })
  }

  public render = () => {
    const { buttonLabel, prompt, title } = this.props
    return (
      <Modal show={this.state.show}>
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
              disabled={() => !passwordValid(this.state.pw)}
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
