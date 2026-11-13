// ABOUTME: Step-by-step scenarios for the TimeoutSequence diagrams.
// ABOUTME: Each click reveals rows, opens or closes timeout windows, and appends history events.
//
// `at` is the click that reveals a thing; click 0 shows the empty diagram and
// the first caption. A slide using a scenario must set `clicks:` to
// `say.length - 1` (checked by snippets/timeoutscenarios/scenarios_test.go).
//
// The history panel shows only events Temporal records, in the compact form
// the Web UI uses (Workflow Task events are left out of Activity scenarios).
// A retried attempt that times out or fails is NOT an event: the Service
// schedules the next attempt internally, and `ActivityTaskStarted` is written
// only when the final attempt closes, with that attempt's number. Ghost rows
// in the panel say so where it happens, because it is the part people expect
// to see and do not.

export type Lane = 0 | 1 | 2
export type Tone = 'ok' | 'fail' | 'plain'

export type Row =
  // `label` is what you would write or see (code, a task, a result); `sub` is
  // the same step in plain English, drawn under the arrow.
  | { kind: 'msg'; from: Lane; to: Lane; label: string; sub?: string; tone?: Tone; at: number; t?: string }
  | { kind: 'note'; lane: Lane; text: string; tone?: Tone; at: number; t?: string }
  | { kind: 'wait'; text: string; at: number; t?: string }

export interface Window {
  label: string
  from: number // first row index the window covers
  to: number // last row index
  col: 0 | 1
  at: number
  close?: { at: number; tone: Tone; text: string }
}

export interface Event {
  name: string
  detail?: string
  tone?: Tone
  at: number
  ghost?: boolean // not a real event: an explanation of what is missing
}

export interface Scenario {
  lanes: [string, string, string]
  code?: string
  rows: Row[]
  windows: Window[]
  events: Event[]
  say: string[]
}

const SERVICE = 'Temporal Service'

export const SCENARIOS: Record<string, Scenario> = {
  'workflow-task': {
    lanes: ['Client', SERVICE, 'Worker'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteWorkflow(OrderWorkflow)', sub: 'app asks Temporal to start it', at: 1, t: '0s' },
      { kind: 'msg', from: 1, to: 2, label: 'run OrderWorkflow code', sub: 'a polling Worker gets the Workflow Task', at: 2, t: '0s' },
      { kind: 'note', lane: 2, text: 'Worker crashes holding the task', tone: 'fail', at: 3, t: '2s' },
      { kind: 'wait', text: 'no response…', at: 4, t: '2–10s' },
      { kind: 'note', lane: 1, text: 'Workflow Task Timeout · reschedule', tone: 'fail', at: 5, t: '10s' },
      { kind: 'msg', from: 1, to: 2, label: 'run OrderWorkflow code', sub: 'another Worker replays the history', at: 6, t: '10s' },
      { kind: 'msg', from: 2, to: 1, label: 'next step: run ChargePayment', sub: 'Worker reports what the Workflow wants', tone: 'ok', at: 7, t: '10s' },
    ],
    windows: [
      { label: 'Workflow Task 10s', from: 1, to: 4, col: 0, at: 2, close: { at: 5, tone: 'fail', text: 'timed out' } },
      { label: 'Workflow Task 10s', from: 5, to: 6, col: 1, at: 6, close: { at: 7, tone: 'ok', text: 'completed' } },
    ],
    events: [
      { name: 'WorkflowExecutionStarted', detail: 'OrderWorkflow', at: 1 },
      { name: 'WorkflowTaskScheduled', detail: 'OrderWorkflow', at: 1 },
      { name: 'WorkflowTaskStarted', detail: 'OrderWorkflow', at: 2 },
      { name: 'WorkflowTaskTimedOut', detail: 'OrderWorkflow · START_TO_CLOSE', tone: 'fail', at: 5 },
      { name: 'retry scheduled: not an Event until it completes', ghost: true, at: 5 },
      { name: 'WorkflowTaskScheduled', detail: 'OrderWorkflow · attempt 2', at: 7 },
      { name: 'WorkflowTaskStarted', detail: 'OrderWorkflow', at: 7 },
      { name: 'WorkflowTaskCompleted', detail: 'OrderWorkflow', tone: 'ok', at: 7 },
      { name: 'ActivityTaskScheduled', detail: 'ChargePayment', at: 7 },
    ],
    say: [
      'A Workflow Task: one chunk of Workflow code a Worker runs, up to the next thing it waits on.',
      'Your app calls client.ExecuteWorkflow. Temporal records WorkflowExecutionStarted and queues the first Workflow Task.',
      'A polling Worker picks it up. It has 10 seconds to report what the Workflow wants to do next.',
      'The Worker process dies while holding the task.',
      'The Service cannot see a crash, so it waits out the timeout.',
      'At 10s the Workflow Task times out, and the Service queues it again.',
      'Another Worker picks it up, replays the history, and carries on.',
      'It answers in time. Only now are the retry\'s Scheduled and Started written. Workflow Tasks retry on their own; no Retry Policy is involved.',
    ],
  },

  'start-to-close': {
    lanes: ['Workflow', SERVICE, 'Worker'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteActivity(ChargePayment)', sub: 'Workflow asks Temporal to run it', at: 1, t: '0s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 1', sub: 'a polling Worker gets the task', at: 2, t: '0s' },
      { kind: 'note', lane: 2, text: 'Worker crashes mid-call', tone: 'fail', at: 3, t: '4s' },
      { kind: 'wait', text: 'no result… the Service cannot see a crash', at: 4, t: '4–10s' },
      { kind: 'note', lane: 1, text: 'Start-To-Close expired · retry in 1s', tone: 'fail', at: 5, t: '10s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 2', sub: 'another Worker gets the task', at: 6, t: '11s' },
      { kind: 'msg', from: 2, to: 1, label: 'ChargePayment returned', sub: 'Worker reports the result', tone: 'ok', at: 7, t: '12s' },
      { kind: 'msg', from: 1, to: 0, label: 'future.Get returns the result', sub: 'Workflow resumes with the result', tone: 'ok', at: 8, t: '12s' },
    ],
    windows: [
      { label: 'Start-To-Close 10s', from: 1, to: 4, col: 0, at: 2, close: { at: 5, tone: 'fail', text: 'expired' } },
      { label: 'Start-To-Close 10s', from: 5, to: 6, col: 1, at: 6, close: { at: 7, tone: 'ok', text: 'closed in 1s' } },
    ],
    events: [
      { name: 'ActivityTaskScheduled', detail: 'ChargePayment', at: 1 },
      { name: 'attempt 1 timed out: a retry, not an Event', ghost: true, at: 5 },
      { name: 'ActivityTaskStarted', detail: 'ChargePayment · attempt 2', at: 7 },
      { name: 'ActivityTaskCompleted', detail: 'ChargePayment · attempt 2', tone: 'ok', at: 7 },
    ],
    say: [
      'One Activity call with a 10-second Start-To-Close Timeout: the limit on a single attempt.',
      'The Workflow calls ExecuteActivity(ChargePayment). That does not run the code: it asks Temporal to, and Temporal records ActivityTaskScheduled.',
      'A Worker is polling the Task Queue. It gets the task, "run ChargePayment, attempt 1", and the 10s clock starts.',
      'The Worker process dies mid-call. Nothing tells the Service.',
      'The Service has no way to see a crash. That is why Start-To-Close is the one timeout you must set.',
      'At 10s the attempt times out, and the Retry Policy schedules attempt 2 after its 1s initial interval.',
      'Another Worker polls and gets attempt 2, with a fresh 10 seconds.',
      'It succeeds. Only now does history record ActivityTaskStarted, and it says attempt 2.',
      'future.Get returns the result in the Workflow. The Workflow never saw attempt 1 at all.',
    ],
  },

  'schedule-to-start': {
    lanes: ['Workflow', 'Task Queue', 'Worker'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteActivity(ChargePayment)', sub: 'Workflow asks Temporal to run it', at: 1, t: '0s' },
      { kind: 'note', lane: 1, text: 'ChargePayment task waits in the queue', at: 2, t: '0s' },
      { kind: 'note', lane: 2, text: 'no Worker polling', tone: 'fail', at: 3, t: '0s' },
      { kind: 'wait', text: 'one minute passes', at: 4, t: '0–60s' },
      { kind: 'note', lane: 1, text: 'Schedule-To-Start expired · never retried', tone: 'fail', at: 5, t: '60s' },
      { kind: 'msg', from: 1, to: 0, label: 'Get returns *ActivityError(*TimeoutError)', sub: 'the Workflow gets the error back', tone: 'fail', at: 6, t: '60s' },
    ],
    windows: [
      { label: 'Schedule-To-Start 1m', from: 1, to: 4, col: 0, at: 2, close: { at: 5, tone: 'fail', text: 'expired' } },
    ],
    events: [
      { name: 'ActivityTaskScheduled', detail: 'ChargePayment', at: 1 },
      { name: 'ActivityTaskTimedOut', detail: 'ChargePayment · SCHEDULE_TO_START', tone: 'fail', at: 5 },
    ],
    say: [
      'Schedule-To-Start: how long an Activity Task may wait in the Task Queue for a Worker.',
      'The Workflow calls ChargePayment. Temporal records ActivityTaskScheduled and puts a task on the Task Queue.',
      'The clock starts while the task waits to be picked up.',
      'No Worker is polling: they are all down, overloaded, or polling a different Task Queue name.',
      'The task just sits there. Nothing is wrong with the Activity code.',
      'At one minute it times out. Schedule-To-Start is never retried.',
      'Get returns an *ActivityError wrapping a *TimeoutError. Usually the fix is more Workers, not this timeout; it is off by default.',
    ],
  },

  heartbeat: {
    lanes: ['Workflow', SERVICE, 'Worker'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteActivity(PackShipment)', sub: 'Workflow asks Temporal to run it', at: 1, t: '0s' },
      { kind: 'msg', from: 1, to: 2, label: 'run PackShipment · attempt 1', sub: 'a polling Worker gets the task', at: 2, t: '0s' },
      { kind: 'msg', from: 2, to: 1, label: 'RecordHeartbeat(ctx, 100)', sub: 'still alive, 100 items packed', tone: 'ok', at: 3, t: '20s' },
      { kind: 'msg', from: 2, to: 1, label: 'RecordHeartbeat(ctx, 200)', sub: 'still alive, 200 items packed', tone: 'ok', at: 4, t: '40s' },
      { kind: 'note', lane: 2, text: 'Worker hangs', tone: 'fail', at: 5, t: '45s' },
      { kind: 'note', lane: 1, text: 'Heartbeat Timeout · retry', tone: 'fail', at: 6, t: '70s' },
      { kind: 'msg', from: 1, to: 2, label: 'run PackShipment · attempt 2', sub: 'with the last heartbeat details: [200]', at: 7, t: '71s' },
      { kind: 'msg', from: 2, to: 1, label: 'PackShipment returned', sub: 'Worker reports the result', tone: 'ok', at: 8, t: '3m' },
    ],
    windows: [
      { label: 'Start-To-Close 1h', from: 1, to: 5, col: 0, at: 2, close: { at: 6, tone: 'plain', text: 'had ~59m left' } },
      { label: 'Heartbeat 30s', from: 3, to: 5, col: 1, at: 4, close: { at: 6, tone: 'fail', text: 'expired' } },
      { label: 'Start-To-Close 1h', from: 6, to: 7, col: 0, at: 7, close: { at: 8, tone: 'ok', text: 'closed' } },
    ],
    events: [
      { name: 'ActivityTaskScheduled', detail: 'PackShipment', at: 1 },
      { name: 'attempt 1 timed out: a retry, not an Event', ghost: true, at: 6 },
      { name: 'ActivityTaskStarted', detail: 'PackShipment · attempt 2', at: 8 },
      { name: 'ActivityTaskCompleted', detail: 'PackShipment · attempt 2', tone: 'ok', at: 8 },
    ],
    say: [
      'A long Activity: an hour of Start-To-Close, and a 30-second Heartbeat Timeout.',
      'The Workflow calls PackShipment.',
      'A polling Worker starts attempt 1. On its own, a crash here would take up to an hour to detect.',
      'The Activity calls activity.RecordHeartbeat, with its progress as the details.',
      'Each Heartbeat resets the 30-second Heartbeat clock.',
      'At 45s the Worker hangs. No more Heartbeats.',
      '30 seconds after the last Heartbeat the attempt times out, about 59 minutes sooner than Start-To-Close.',
      'Attempt 2 starts with the last Heartbeat details, and resumes from item 200.',
      'It completes. History shows one Started (attempt 2) and one Completed.',
    ],
  },

  'schedule-to-close': {
    lanes: ['Workflow', SERVICE, 'Worker'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteActivity(ChargePayment)', sub: 'Workflow asks Temporal to run it', at: 1, t: '0s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 1', sub: 'the payment API hangs', at: 2, t: '0s' },
      { kind: 'note', lane: 1, text: 'Start-To-Close expired · retry in 1s', tone: 'fail', at: 3, t: '3s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 2', sub: 'hangs again', at: 4, t: '4s' },
      { kind: 'note', lane: 1, text: 'Start-To-Close expired · retry in 2s', tone: 'fail', at: 5, t: '7s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 3', sub: 'one second of budget left', at: 6, t: '9s' },
      { kind: 'note', lane: 1, text: 'Schedule-To-Close expired · no more attempts', tone: 'fail', at: 7, t: '10s' },
      { kind: 'msg', from: 1, to: 0, label: 'Get returns *ActivityError(*TimeoutError)', sub: 'the Workflow gets the error back', tone: 'fail', at: 8, t: '10s' },
    ],
    windows: [
      { label: 'Schedule-To-Close 10s', from: 0, to: 6, col: 0, at: 1, close: { at: 7, tone: 'fail', text: 'expired' } },
      { label: 'S-T-C 3s', from: 1, to: 2, col: 1, at: 2, close: { at: 3, tone: 'fail', text: '' } },
      { label: 'S-T-C 3s', from: 3, to: 4, col: 1, at: 4, close: { at: 5, tone: 'fail', text: '' } },
      { label: 'S-T-C 3s', from: 5, to: 6, col: 1, at: 6, close: { at: 7, tone: 'fail', text: 'cut short' } },
    ],
    events: [
      { name: 'ActivityTaskScheduled', detail: 'ChargePayment', at: 1 },
      { name: 'attempts 1 and 2 timed out: retries, not Events', ghost: true, at: 5 },
      { name: 'ActivityTaskStarted', detail: 'ChargePayment · attempt 3', at: 7 },
      { name: 'ActivityTaskTimedOut', detail: 'ChargePayment · SCHEDULE_TO_CLOSE', tone: 'fail', at: 7 },
    ],
    say: [
      'Schedule-To-Close: a budget for the whole Activity, every attempt and every wait between them.',
      'The Workflow calls ChargePayment. The 10s budget starts now, before any Worker has the task.',
      'Attempt 1 calls the payment API, which hangs.',
      'Its 3 seconds of Start-To-Close run out. Retry in 1s.',
      'Attempt 2, with a fresh 3 seconds. It hangs too.',
      'Retry again, after 2s this time: the interval doubles.',
      'Attempt 3 starts at 9s, with only one second of budget left.',
      'At 10s the whole Activity times out, whatever the Retry Policy would still allow.',
      'Get returns an *ActivityError wrapping a *TimeoutError. Bound retries with this, not with Maximum Attempts.',
    ],
  },

  'max-attempts': {
    lanes: ['Workflow', SERVICE, 'Worker → Payment API'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteActivity(ChargePayment)', at: 1, t: '0s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 1', at: 2, t: '0s' },
      { kind: 'msg', from: 2, to: 1, label: 'ChargePayment returned an error', sub: 'the API returned 503', tone: 'fail', at: 3, t: '0.2s' },
      { kind: 'wait', text: 'retry 1 of 2 · wait 1s (Initial Interval)', at: 4, t: '0.2–1.2s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 2', at: 5, t: '1.2s' },
      { kind: 'msg', from: 2, to: 1, label: 'ChargePayment returned an error', sub: '503 again', tone: 'fail', at: 6, t: '1.4s' },
      { kind: 'wait', text: 'retry 2 of 2 · wait 2s (1s × Backoff 2.0)', at: 7, t: '1.4–3.4s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 3', at: 8, t: '3.4s' },
      { kind: 'msg', from: 2, to: 1, label: 'ChargePayment returned an error', sub: '503 a third time', tone: 'fail', at: 9, t: '3.6s' },
      { kind: 'note', lane: 1, text: 'Maximum Attempts reached · no more retries', tone: 'fail', at: 10, t: '3.6s' },
      { kind: 'msg', from: 1, to: 0, label: 'Get returns *ActivityError(*ApplicationError)', sub: 'the Workflow gets the error back', tone: 'fail', at: 11, t: '3.6s' },
    ],
    windows: [],
    events: [
      { name: 'ActivityTaskScheduled', detail: 'ChargePayment', at: 1 },
      { name: 'failed attempts are retries, not Events', ghost: true, at: 4 },
      { name: 'ActivityTaskStarted', detail: 'ChargePayment · attempt 3', at: 10 },
      { name: 'ActivityTaskFailed', detail: 'ChargePayment · MAXIMUM_ATTEMPTS_REACHED', tone: 'fail', at: 10 },
    ],
    say: [
      'A Retry Policy capped at three attempts, against a payment API that keeps returning 503.',
      'The Workflow calls ChargePayment with RetryPolicy{MaximumAttempts: 3}.',
      'A polling Worker runs attempt 1, which calls the API.',
      'The API fails, and ChargePayment returns an error.',
      'Retry 1 of 2: wait the Initial Interval, 1 second.',
      'Attempt 2.',
      'Fails again.',
      'Retry 2 of 2: the wait doubles, by the Backoff Coefficient, to 2 seconds.',
      'Attempt 3.',
      'Fails a third time.',
      'Maximum Attempts reached. History now records the last attempt, and why it stopped.',
      'Get returns an *ActivityError, and the Workflow decides what to do: alert, compensate, or escalate.',
    ],
  },

  'non-retryable': {
    lanes: ['Workflow', SERVICE, 'Worker → Payment API'],
    rows: [
      { kind: 'msg', from: 0, to: 1, label: 'ExecuteActivity(ChargePayment)', sub: 'Workflow asks Temporal to run it', at: 1, t: '0s' },
      { kind: 'msg', from: 1, to: 2, label: 'run ChargePayment · attempt 1', sub: 'a polling Worker gets the task', at: 2, t: '0s' },
      { kind: 'msg', from: 2, to: 1, label: 'return NewNonRetryableApplicationError(...)', sub: 'errType: "CardDeclined"', tone: 'fail', at: 3, t: '0.3s' },
      { kind: 'note', lane: 1, text: 'non-retryable · stop, whatever the policy', tone: 'fail', at: 4, t: '0.3s' },
      { kind: 'msg', from: 1, to: 0, label: 'Get returns *ActivityError(*ApplicationError)', sub: 'appErr.Type() == "CardDeclined"', tone: 'fail', at: 5, t: '0.3s' },
    ],
    windows: [],
    events: [
      { name: 'ActivityTaskScheduled', detail: 'ChargePayment', at: 1 },
      { name: 'ActivityTaskStarted', detail: 'ChargePayment · attempt 1', at: 4 },
      { name: 'ActivityTaskFailed', detail: 'ChargePayment · NON_RETRYABLE_FAILURE', tone: 'fail', at: 4 },
    ],
    say: [
      'The same charge, but the card is declined. Retrying will never help.',
      'The Workflow calls ChargePayment.',
      'A polling Worker runs attempt 1, which calls the payment API.',
      'ChargePayment returns temporal.NewNonRetryableApplicationError with type "CardDeclined".',
      'The Service stops at once, with retries left in the policy.',
      'Get returns an *ActivityError whose cause has Type() "CardDeclined", one attempt and 0.3 seconds later.',
    ],
  },
}
