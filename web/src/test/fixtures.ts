import type { Report, SpecStatus } from '../types';

export function spec(partial: Partial<SpecStatus> & { id: string }): SpecStatus {
  return {
    title: partial.id,
    path: `specs/misc/${partial.id}.md`,
    tests: [],
    covered: false,
    coversOk: true,
    draft: false,
    ...partial,
  };
}

// A report exercising every visual state: covered, uncovered (error),
// draft (warning), and covers-unmatched (warning on a covered spec).
export const mixedReport: Report = {
  ok: false,
  testFiles: 6,
  specs: [
    spec({
      id: 'auth-login',
      title: 'A user can log in',
      path: 'specs/auth/login.md',
      covered: true,
      coversOk: true,
      tests: ['server/auth_test.go', 'web/e2e/auth.spec.ts'],
      covers: ['server/auth.go'],
      body: '## Behaviour\n\nValid credentials return a session.\n',
    }),
    spec({
      id: 'auth-logout',
      title: 'Logout revokes the session',
      path: 'specs/auth/logout.md',
      covered: true,
      coversOk: false, // covers-unmatched → warning
      tests: ['server/auth_test.go'],
      covers: ['server/ghost.go'],
    }),
    spec({
      id: 'tasks-create',
      title: 'A user can create a task',
      path: 'specs/tasks/create.md',
      covered: false, // uncovered → error
      draft: false,
      tests: [],
    }),
    spec({
      id: 'sharing-permissions',
      title: 'Collaborators are read-only until promoted',
      path: 'specs/sharing/permissions.md',
      status: 'draft',
      draft: true,
      covered: false, // uncovered draft → warning, not error
      tests: [],
    }),
  ],
  findings: [
    {
      severity: 'error',
      rule: 'uncovered-spec',
      spec: 'tasks-create',
      file: 'specs/tasks/create.md',
      message: 'no test references spec:tasks-create',
    },
    {
      severity: 'warning',
      rule: 'covers-unmatched',
      spec: 'auth-logout',
      file: 'specs/auth/logout.md',
      message: 'covers entry server/ghost.go matches no file',
    },
    {
      severity: 'warning',
      rule: 'uncovered-draft',
      spec: 'sharing-permissions',
      file: 'specs/sharing/permissions.md',
      message: 'draft spec has no covering test yet',
    },
  ],
};

// A report with an ingested test run: one passing, one FAILING (covered but
// red), one skipped, and one uncovered — exercising the result states.
export const resultsReport: Report = {
  ok: true, // traceability passes; the failing test is what should flip the verdict
  testFiles: 4,
  hasResults: true,
  specs: [
    spec({
      id: 'auth-login',
      title: 'A user can log in',
      path: 'specs/auth/login.md',
      covered: true,
      result: 'passed',
      tests: ['server/auth_test.go'],
      refs: [{ file: 'server/auth_test.go', line: 5, test: 'TestLogin', status: 'passed' }],
    }),
    spec({
      id: 'auth-logout',
      title: 'Logout revokes the session',
      path: 'specs/auth/logout.md',
      covered: true,
      result: 'failed',
      tests: ['server/auth_test.go'],
      refs: [{ file: 'server/auth_test.go', line: 12, test: 'TestLogout', status: 'failed' }],
    }),
    spec({
      id: 'tasks-toggle',
      title: 'Toggling a task persists',
      path: 'specs/tasks/toggle.md',
      covered: true,
      result: 'skipped',
      tests: ['server/tasks_test.go'],
      refs: [{ file: 'server/tasks_test.go', line: 3, test: 'TestToggle', status: 'skipped' }],
    }),
    spec({
      id: 'tasks-create',
      title: 'A user can create a task',
      path: 'specs/tasks/create.md',
      covered: false,
      tests: [],
    }),
  ],
  findings: [
    {
      severity: 'error',
      rule: 'uncovered-spec',
      spec: 'tasks-create',
      file: 'specs/tasks/create.md',
      message: 'no test references spec:tasks-create',
    },
  ],
};

export const cleanReport: Report = {
  ok: true,
  testFiles: 3,
  specs: [
    spec({
      id: 'auth-login',
      title: 'A user can log in',
      path: 'specs/auth/login.md',
      covered: true,
      coversOk: true,
      tests: ['server/auth_test.go'],
    }),
  ],
  findings: [],
};
