#!/usr/bin/env python3
"""
claude_transcriptor - Convert a Claude Code session JSONL file to a Markdown interview transcript.

Usage: claude_transcriptor <session.jsonl> [output.md]

Session files are found in:
  ~/.claude/projects/<project-slug>/<session-id>.jsonl
"""

import json
import sys
import os


def format_ask_question(item):
    """Format an AskUserQuestion tool use as a list of options."""
    q_input = item.get('input', {})
    questions = q_input.get('questions', [])
    parts = []
    for q in questions:
        q_text = q.get('question', '')
        options = q.get('options', [])
        formatted = f"**{q_text}**\n\n"
        for opt in options:
            label = opt.get('label', '')
            desc = opt.get('description', '')
            formatted += f"- **{label}**: {desc}\n"
        parts.append(formatted)
    return parts


def parse_session(session_file):
    transcript = []

    # Load all entries for two-pass processing
    entries = []
    with open(session_file, 'r') as f:
        for line in f:
            try:
                entries.append(json.loads(line))
            except Exception:
                entries.append(None)

    # Pre-pass: collect ExitPlanMode tool use IDs
    exit_plan_ids = set()
    for entry in entries:
        if entry and entry.get('type') == 'assistant':
            msg = entry.get('message', {})
            for item in msg.get('content', []):
                if isinstance(item, dict) and item.get('name') == 'ExitPlanMode':
                    exit_plan_ids.add(item['id'])

    for entry in entries:
        try:
            if not entry:
                continue
            ts = entry.get('timestamp', '')

            if entry.get('type') == 'user':
                msg = entry.get('message')

                # Regular typed user messages
                if isinstance(msg, dict):
                    content = msg.get('content', '')
                    if isinstance(content, str) and content.strip() and not content.startswith('<'):
                        content = content.replace('\\\n', '\n')
                        transcript.append(('Josh', content, ts))
                    elif isinstance(content, list):
                        for item in content:
                            if (isinstance(item, dict) and
                                    item.get('type') == 'tool_result' and
                                    item.get('tool_use_id') in exit_plan_ids):
                                is_err = item.get('is_error', False)
                                raw = item.get('content', '')
                                if is_err:
                                    feedback = ''
                                    if 'the user said:\n' in raw:
                                        feedback = raw.split('the user said:\n', 1)[-1].strip()
                                    if feedback:
                                        transcript.append(('Josh', f'**[Plan rejected]**\n\n{feedback}', ts))
                                    else:
                                        transcript.append(('Josh', '**[Plan rejected]**', ts))
                                else:
                                    transcript.append(('Josh', '**[Plan accepted]**', ts))

                    # Tool results: successful answers to AskUserQuestion
                    result = entry.get('toolUseResult')
                    if isinstance(result, dict):
                        answers = result.get('answers', {})
                        if isinstance(answers, dict) and answers:
                            parts = list(answers.values())
                            combined = '\n\n'.join(parts)
                            transcript.append(('Josh', combined, ts))

                    # Tool results: rejected tool uses (e.g. user clicked "clarify")
                    elif isinstance(result, str) and result.startswith('Error:'):
                        msg_content = msg.get('content', []) if isinstance(msg, dict) else []
                        if isinstance(msg_content, list):
                            for item in msg_content:
                                if isinstance(item, dict) and item.get('type') == 'tool_result' and item.get('is_error'):
                                    if 'wants to clarify' in result:
                                        transcript.append(('Josh', '<requested clarification>', ts))
                                    break

            elif entry.get('type') == 'assistant':
                msg = entry.get('message')
                if isinstance(msg, dict):
                    for item in msg.get('content', []):
                        if not isinstance(item, dict):
                            continue
                        if item.get('type') == 'text':
                            text = item.get('text', '')
                            if text.strip():
                                text = text.replace('\\\n', '\n')
                                transcript.append(('Claude', text, ts))
                        elif item.get('type') == 'tool_use' and item.get('name') == 'AskUserQuestion':
                            for formatted in format_ask_question(item):
                                transcript.append(('Claude', formatted, ts))
                        elif item.get('type') == 'tool_use' and item.get('name') == 'ExitPlanMode':
                            plan = item.get('input', {}).get('plan', '')
                            if plan.strip():
                                transcript.append(('Claude', f'**[Proposed Plan]**\n\n{plan}', ts))

            elif entry.get('type') == 'progress':
                data_msg = entry.get('data', {}).get('message', {})
                if data_msg.get('type') == 'user':
                    for item in data_msg.get('message', {}).get('content', []):
                        if isinstance(item, dict) and item.get('type') == 'text':
                            text = item.get('text', '')
                            if text.strip() and not text.startswith('<'):
                                text = text.replace('\\\n', '\n')
                                transcript.append(('Claude', f'**[To subagent]** {text}', ts))
                elif data_msg.get('type') == 'assistant':
                    for item in data_msg.get('message', {}).get('content', []):
                        if isinstance(item, dict) and item.get('type') == 'text':
                            text = item.get('text', '')
                            if text.strip():
                                text = text.replace('\\\n', '\n')
                                transcript.append(('Claude', text, ts))

        except Exception:
            continue

    transcript.sort(key=lambda x: x[2] if x[2] else '')
    return transcript


def write_transcript(transcript, output_file):
    output = "# Interview Transcript: Claude Code Session\n\n"
    for speaker, message, _ in transcript:
        output += f"**{speaker}:** {message}\n\n"
    with open(output_file, 'w') as f:
        f.write(output)


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        sys.exit(1)

    session_file = sys.argv[1]
    if not os.path.exists(session_file):
        print(f"Error: file not found: {session_file}", file=sys.stderr)
        sys.exit(1)

    output_file = sys.argv[2] if len(sys.argv) > 2 else 'transcript.md'

    transcript = parse_session(session_file)
    write_transcript(transcript, output_file)
    print(f"Transcript written with {len(transcript)} exchanges to {output_file}")


if __name__ == '__main__':
    main()
