type SlackCookie = {
  d: string
}

type SlackLocalConfig = {
  teams: {
    [key: string]: {
      id: string
      token: string
    }
  }
}

type SlackTeamInfo = {
  team: {
    id: string
    name: string
    domain: string
  }
}

async function getCookie(): Promise<SlackCookie> {
  const process = new Deno.Command('extract-chrome-storage', {
    args: [
      'cookie',
      '--browser',
      'slack',
      '--app-store',
      '--domain',
      '.slack.com',
    ],
    stdout: 'piped',
  }).spawn()
  const { stdout } = await process.output()
  return JSON.parse(new TextDecoder().decode(stdout)) as SlackCookie
}

async function getLocalConfig(): Promise<SlackLocalConfig> {
  const process = new Deno.Command('extract-chrome-storage', {
    args: [
      'local-storage',
      '--browser',
      'slack',
      '--app-store',
      '--domain',
      'app.slack.com',
      '--key',
      'localConfig_v2',
    ],
    stdout: 'piped',
  }).spawn()
  const { stdout } = await process.output()
  return JSON.parse(new TextDecoder().decode(stdout)) as SlackLocalConfig
}

async function getTeamInfo(
  teamId: string,
  options: {
    cookie: SlackCookie
    localConfig: SlackLocalConfig
  },
): Promise<SlackTeamInfo> {
  const response = await fetch('https://slack.com/api/team.info', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Cookie': `d=${options.cookie.d}`,
      'Authorization': `Bearer ${options.localConfig.teams[teamId].token}`,
    },
    body: JSON.stringify({ team: teamId }),
  })

  if (!response.ok) {
    throw new Error(`Failed to fetch team info: ${response.statusText}`)
  }

  return await response.json() as SlackTeamInfo
}

async function main() {
  const cookie = await getCookie()
  const localConfig = await getLocalConfig()

  const teams = Object.entries(localConfig.teams)
    .map(([_, teamInfo]) => teamInfo)

  for (const team of teams) {
    const teamInfo = await getTeamInfo(team.id, { cookie, localConfig })
    console.log(teamInfo)
  }
}

await main()
