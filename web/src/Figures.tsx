import React, { useCallback, useState } from 'react'

import {
  AppBar,
  Box,
  Card,
  CardContent,
  Link,
  Paper,
  Tab,
  Tooltip,
  Typography,
} from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'
import { TabContext, TabList, TabPanel } from '@material-ui/lab'

import { FigureData, UserData } from './types'

interface Props {
  diedToday: FigureData[]
  outlived: FigureData[] | null
}

const useStyles = (theme: Theme) =>
  makeStyles({
    card: {
      display: 'inline-block',
      width: '14em',
      margin: '6px',
      textAlign: 'center',
      verticalAlign: 'top',
    },
  })

export const Figures = (props: Props) => {
  const { diedToday, outlived } = props

  const [activeTab, setActiveTab] = useState(
    outlived ? 'you-outlived' : 'died-today'
  )

  const theme = useTheme()
  const classes = useStyles(theme)()

  const outlivedTab = (
    <Tab
      value='you-outlived'
      disabled={!outlived}
      label='You have recently outlived'
    />
  )

  return (
    <Box alignItems='center' justifyContent='center' textAlign='center'>
      <TabContext value={activeTab}>
        <AppBar position='static'>
          <TabList
            onChange={(event: React.ChangeEvent<{}>, newValue: string) =>
              setActiveTab(newValue)
            }
          >
            <Tab value='died-today' label='Died on this date' />
            {outlived ? (
              outlivedTab
            ) : (
              <Tooltip title='Log in to see whom you’ve recently outlived.'>
                <span>{outlivedTab}</span>
              </Tooltip>
            )}
          </TabList>
        </AppBar>
        <TabPanel value='died-today'>{renderFigs(diedToday, classes)}</TabPanel>
        <TabPanel value='you-outlived'>
          {outlived ? renderFigs(outlived, classes) : null}
        </TabPanel>
      </TabContext>
    </Box>
  )
}

// xxx figure out the right type for classes
const renderFigs = (figs: FigureData[], classes: any) => {
  return (
    <>
      {figs.map((fig: FigureData) => (
        <Card className={classes.card} key={fig.href}>
          <CardContent>
            <Link
              className='figure'
              target='_blank'
              rel='noopener noreferrer'
              href={fig.href}
            >
              {fig.imgSrc ? (
                <>
                  <img
                    className={classes.image}
                    src={fig.imgSrc}
                    alt={fig.imgAlt}
                  />
                  <br />
                </>
              ) : undefined}
              {fig.name}
            </Link>
            <br />
            {fig.desc}
            {fig.desc ? <br /> : null}
            {fig.born}&mdash;{fig.died}
            <br />({fig.yearsDaysAlive})
          </CardContent>
        </Card>
      ))}
    </>
  )
}
