import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'
import DatePicker from 'react-date-picker'
import Toggle from 'react-toggle'

import { LogoutButton } from './LogoutButton'
import { PasswordDialog } from './Password'
import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'

interface Props {
  onLogin: (user: UserData) => void
  user?: UserData
}

interface State {
  birthDate: Date
  email?: string
  enteringBirthdate?: boolean
  enteringPassword?: boolean
  password?: string
  receivingMail?: boolean
  signingUp?: boolean
}

export class User extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      birthDate: new Date(),
    }
    console.log(`xxx props.user ${JSON.stringify(props.user)}`)
  }

  private setActive = async (active: boolean) => {
    if (!this.props.user) {
      console.log(`xxx setActive short-circuiting`)
      return
    }
    const { csrf } = this.props.user
    const resp = await post('/s/setactive', {
      csrf,
      active,
    })
    // xxx check resp
    console.log(`xxx setActive, setting receivingMail to ${active}`)
    this.setState({ receivingMail: active })
  }

  private onPassword = async (pw: string) => {
    const { email } = this.state
    this.setState({
      enteringBirthdate: this.state.signingUp,
      password: pw,
    })
    if (this.state.signingUp) {
      return
    }
    const resp = await post('/s/login', {
      email,
      password: pw,
      tzname: tzname(),
    })
    // xxx check for error
    const user = (await resp.json()) as UserData
    this.props.onLogin(user)
  }

  private onBirthDate = async (ev: React.FormEvent<HTMLFormElement>) => {
    ev.preventDefault()
    const { birthDate } = this.state
    if (!birthdateValid(birthDate)) {
      return
    }
    const { email, password } = this.state
    const resp = await post('/s/signup', {
      email,
      password,
      born: datestr(birthDate),
      tzname: tzname(),
    })
    const user = (await resp.json()) as UserData
    this.props.onLogin(user)
  }

  public componentDidMount = () => {
    const { user } = this.props

    console.log(`xxx componentDidMount: user is ${JSON.stringify(user)}`)

    if (!user) {
      return
    }
    const { active, verified } = user
    console.log(
      `xxx componentDidMount: setting receivingMail to ${verified && active}`
    )
    this.setState({ receivingMail: verified && active })
  }

  public shouldComponentUpdate = (
    nextProps: Props,
    nextState: State,
    nextContent: any
  ) => {
    const superShould = super.shouldComponentUpdate
    if (superShould && superShould(nextProps, nextState, nextContent)) {
      console.log(`xxx super.shouldComponentUpdate says yes`)
      return true
    }
    if (!nextProps.user !== !this.props.user) {
      console.log(`xxx super.shouldComponentUpdate says no but I say yes [1]`)
      return true
    }
    if (!nextState.receivingMail !== !this.state.receivingMail) {
      console.log(`xxx super.shouldComponentUpdate says no but I say yes [2]`)
      return true
    }
    return false
  }

  public render = () => {
    const { user } = this.props

    console.log(
      `xxx User render, !!user is ${!!user}, receivingMail is ${
        this.state.receivingMail
      }`
    )

    if (user) {
      const { active, csrf, email, verified } = user
      return (
        <div>
          <div>
            Logged in as {email}.
            <LogoutButton csrf={csrf} />
          </div>
          <div>
            <label htmlFor='active'>
              <span>Receive Outlived mail?</span>
              <Toggle
                id='active'
                checked={!!this.state.receivingMail}
                disabled={!user.verified}
                onChange={ev => {
                  console.log(`xxx flipping toggle, ev.target.checked is ${ev.target.checked}`)
                  this.setActive(ev.target.checked)
                }}
              />
            </label>
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
                  this.setState({ birthDate: date }) // xxx hack
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
      </>
    )
  }
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
