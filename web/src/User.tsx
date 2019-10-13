import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'
import DatePicker from 'react-date-picker'
import Toggle from 'react-toggle'
import { Label, Popup } from 'semantic-ui-react'

import { doAlert } from './Alert'
import { PasswordDialog } from './Password'
import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'

interface LoggedInProps {
  user: UserData
}

interface LoggedInState {
  receivingMail: boolean
  reverified: boolean
}

export class LoggedInUser extends React.Component<
  LoggedInProps,
  LoggedInState
> {
  constructor(props: LoggedInProps) {
    super(props)
    const { user } = props
    this.state = {
      receivingMail: user.verified && user.active,
      reverified: false,
    }
  }

  private setActive = async (active: boolean) => {
    const { csrf } = this.props.user
    try {
      await post('/s/setactive', {
        csrf,
        active,
      })
      this.setState({ receivingMail: active })
    } catch (error) {
      doAlert('Error setting e-mail preference. Please try again.')
    }
  }

  private reverify = async () => {
    const { csrf } = this.props.user
    try {
      await post('/s/reverify', { csrf })
      this.setState({ reverified: true })
    } catch (error) {
      doAlert(
        'Error resending verification message. Please try again in a moment.'
      )
    }
  }

  public render = () => {
    const { user } = this.props
    const { csrf, email, verified } = user

    return (
      <div className='user logged-in'>
        <div>
          Logged in as {email}.
          <form method='POST' action='/s/logout'>
            <input type='hidden' name='csrf' value={csrf} />
            <button type='submit'>Log out</button>
          </form>
        </div>
        <div>
          <Popup
            content='Up to one message per day showing the notable figures you’ve just outlived.'
            position='left center'
            trigger={
              <Label>
                Receive Outlived mail?
                <br />
                <Toggle
                  id='active'
                  checked={!!this.state.receivingMail}
                  disabled={!user.verified}
                  onChange={ev => this.setActive(ev.target.checked)}
                />
              </Label>
            }
          />

          {!verified &&
            (this.state.reverified ? (
              <span id='reverified'>
                Check your e-mail for a verification message from Outlived.
              </span>
            ) : (
              <span id='unconfirmed'>
                You have not yet confirmed your e-mail address.
                <button id='resend-button' onClick={this.reverify}>
                  Resend verification
                </button>
              </span>
            ))}
        </div>
      </div>
    )
  }
}

interface LoggedOutProps {
  onLogin: (user: UserData) => void
}

interface LoggedOutState {
  birthDate: Date
  email?: string
  enteringBirthdate?: boolean
  enteringPassword?: boolean
  password?: string
  signingUp?: boolean
}

export class LoggedOutUser extends React.Component<
  LoggedOutProps,
  LoggedOutState
> {
  public state: LoggedOutState = { birthDate: new Date() }

  private onPassword = async (pw: string) => {
    const { email } = this.state
    this.setState({
      enteringBirthdate: this.state.signingUp,
      password: pw,
    })
    if (this.state.signingUp) {
      return
    }

    try {
      const resp = await post('/s/login', {
        email,
        forgot: false,
        password: pw,
        tzname: tzname(),
      })
      const user = (await resp.json()) as UserData
      this.props.onLogin(user)
    } catch (error) {
      doAlert('Login failed')
    }
  }

  private onForgotPassword = async () => {
    const { email } = this.state
    if (!emailValid(email)) {
      return
    }
    try {
      await post('/s/login', {
        email,
        forgot: true,
        tzname: tzname(),
      })
      doAlert('Check your e-mail for a password-reset message.')
    } catch (error) {
      doAlert(
        'Error sending password-reset e-mail. Please try again in a moment.'
      )
    }
  }

  private onBirthDate = async (ev: React.FormEvent<HTMLFormElement>) => {
    ev.preventDefault()
    const { birthDate } = this.state
    if (!birthdateValid(birthDate)) {
      return
    }
    const { email, password } = this.state
    try {
      const resp = await post('/s/signup', {
        email,
        password,
        born: datestr(birthDate),
        tzname: tzname(),
      })
      const user = (await resp.json()) as UserData
      this.props.onLogin(user)
    } catch (error) {
      doAlert(`Error creating account. Please try again.`)
    }
  }

  public render = () => (
    <div className='user logged-out'>
      <p>Log in to see whom you’ve recently outlived.</p>

      <div>
        <Label>
          E-mail address{' '}
          <input
            type='email'
            id='email'
            name='email'
            onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
              this.setState({ email: ev.target.value })
            }
          />
        </Label>
        <br />
        <button
          onClick={() => this.setState({ enteringPassword: true })}
          disabled={!emailValid(this.state.email)}
        >
          Log in
        </button>
        <button
          onClick={() =>
            this.setState({ enteringPassword: true, signingUp: true })
          }
          disabled={!emailValid(this.state.email)}
        >
          Sign up
        </button>
      </div>

      <PasswordDialog
        prompt={this.state.signingUp ? 'Choose password' : 'Enter password'}
        show={() => !!this.state.enteringPassword}
        onClose={() => this.setState({ enteringPassword: false })}
        onForgot={this.state.signingUp ? undefined : this.onForgotPassword}
        onSubmit={this.onPassword}
      />

      <Modal
        onHide={() =>
          this.setState({ enteringBirthdate: false, signingUp: false })
        }
        show={this.state.enteringBirthdate}
      >
        <Modal.Header closeButton>
          <ModalTitle>Birth date</ModalTitle>
        </Modal.Header>
        <ModalBody>
          <form onSubmit={this.onBirthDate}>
            <DatePicker
              onChange={(d: Date | Date[]) => {
                const date = d as Date // xxx hack
                this.setState({ birthDate: date })
              }}
              value={this.state.birthDate}
            />
            <button
              type='submit'
              disabled={!birthdateValid(this.state.birthDate)}
            >
              Submit
            </button>
          </form>
        </ModalBody>
      </Modal>
    </div>
  )
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp?: string) => {
  return inp && /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}

const yearMillis =
  365 /* days */ * 24 /* hours */ * 60 /* mins */ * 60 /* secs */ * 1000

const birthdateValid = (d: Date) => {
  const now = new Date()
  const diff = now.valueOf() - d.valueOf()
  if (diff < 13 * yearMillis) {
    return false
  }
  if (diff > 200 * yearMillis) {
    return false
  }
  return true
}

const datestr = (d: Date) => {
  const y = d.getFullYear()
  const m = 1 + d.getMonth() // come on javascript
  const day = d.getDate() // come ON, javascript
  return y + '-' + ('0' + m).substr(-2) + '-' + ('0' + day).substr(-2)
}
