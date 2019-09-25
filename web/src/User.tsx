import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'

import { LogoutButton } from './LogoutButton'
import { ReceiveMailCheckbox } from './ReceiveMailCheckbox'
import { UserData } from './types'
import { tzname } from './tz'

interface Props {
  onLogin: (user: UserData) => void
  user?: UserData
}

interface State {
  born?: string
  email?: string
  password?: string
  loggingIn?: boolean
}

export class User extends React.Component<Props, State> {
  public state: State = {}

  private login = async () => {
    const { email, password } = this.state

    this.setState({ loggingIn: false })

    const resp = await fetch('/s/login', {
      method: 'POST',
      credentials: 'same-origin',
      body: JSON.stringify({
        email,
        password,
        tzname: tzname(),
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    })
    const user = (await resp.json()) as UserData
    this.props.onLogin(user)
  }

  private signup = () => {
    const { born, email, password } = this.state

    fetch('/s/signup', {
      method: 'POST',
      credentials: 'same-origin',
      body: JSON.stringify({
        email,
        password,
        born,
        tzname: tzname(),
      }),
      headers: {
        'Content-Type': 'application/json',
      },
    })
  }

  public render = () => {
    const { user } = this.props

    if (user) {
      const { csrf, email } = user
      return (
        <div>
          <div>
            Signed in as {email}.
            <LogoutButton csrf={csrf} />
          </div>
          <div>
            <ReceiveMailCheckbox csrf={csrf} user={user} />
          </div>
        </div>
      )
    }

    return (
      <>
        <p>Log in to see whom you’ve recently outlived.</p>

        <div>
          <label htmlFor='email'>E-mail address</label>
          <input
            type='email'
            id='email'
            name='email'
            onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
              this.setState({ email: ev.target.value })
            }
          />
          <button
            onClick={() => this.setState({ loggingIn: true })}
            disabled={!emailValid(this.state.email)}
          >
            Log in
          </button>
        </div>

        {this.state.loggingIn && (
          <Modal show={true} onExit={() => this.setState({ loggingIn: false })}>
            <Modal.Header closeButton>
              <ModalTitle>Password</ModalTitle>
            </Modal.Header>

            <ModalBody>
              <label htmlFor='password'>Password for {this.state.email}</label>
              <input
                type='password'
                id='password'
                name='password'
                onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
                  this.setState({ password: ev.target.value })
                }
              />
              <button onClick={this.login} disabled={!this.state.password}>
                Log in
              </button>
            </ModalBody>
          </Modal>
        )}
      </>
    )
  }
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp?: string) => {
  return inp && /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}

const passwordValid = (inp: string) => {
  return !!inp
}
