export function replaceCommandTokens(command: string, wordsToReplace: { [key: string]: string }): string {
  let cmd = command;

  if (cmd.includes('-ArgumentList')) {

    const args = Object.values(wordsToReplace)
      .filter(v => v && v.trim().length > 0)
      .map(v => `'${v.trim()}'`)
      .join(', ');

    cmd = cmd.replace(
      /-ArgumentList\s+(['"].*?['"])(?=\s+-|$)/,
      `-ArgumentList ${args}`
    );
  } else {
    const match = cmd.match(/"(.*)"/);
    if (match) {
      const original = match[1];
      const parts = original.split(/\s+/);
      const fixedCommand = parts[0];
      const args = Object.entries(wordsToReplace)
        .filter(([_, v]) => v && v.trim().length > 0)
        .map(([_, v]) => v.trim())
        .join(' ');

      cmd = cmd.replace(/"(.*)"/, `"${fixedCommand} ${args}"`);
    }
  }

  cmd = cmd.replace(/\s+/g, ' ').trim();

  return cmd;
}
