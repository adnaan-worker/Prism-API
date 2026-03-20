export function formatGreeting(name: string): string {
  return `Hello, ${name}`;
}

export const demoUser = {
  id: 1,
  name: 'Prism',
};

export const greetingMessage = formatGreeting(demoUser.name);
