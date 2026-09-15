import type { Format } from './types';

// Mirrors internal/parse/parse.go — keep the two in sync.
// NPS markers cover Event Viewer text (6272/6273/6274), the Event Viewer
// XML view, NPS System-log events 13/18, and the IAS/DTS accounting log-file
// formats. `IAS` is word-bounded: unanchored it matched "tobias" or "alias"
// inside FreeRADIUS output and flipped the format to NPS.
const npsRe =
  /Network Policy Server|Reason Code:|Event ID:\s*627[234]|<EventID>627[234]<\/EventID>|Name="ReasonCode"|\bIAS\b|<Reason-Code data_type=|<Packet-Type data_type=|invalid RADIUS client IP address|Message-Authenticator attribute that is not valid/i;
const frRe =
  /FreeRADIUS|radiusd|Access-(Request|Accept|Reject|Challenge)|rlm_|eap_(peap|tls|ttls|md5)|\(\d+\) /i;

// NPS is checked first: NPS event text never contains FreeRADIUS markers,
// while pasted mixtures should lean NPS.
export function detect(text: string): Format {
  if (npsRe.test(text)) return 'nps';
  if (frRe.test(text)) return 'freeradius';
  return 'unknown';
}

const npsFields: Record<string, string> = {
  'Reason Code': 'reason_code',
  Reason: 'reason',
  'Authentication Type': 'authentication_type',
  'EAP Type': 'eap_type',
  'Network Policy Name': 'network_policy',
  'Connection Request Policy Name': 'connection_request_policy',
  'NAS Port-Type': 'nas_port_type',
  'Event ID': 'event_id',
};

const errorLineRe = /\bERROR\b/i;
const sentReplyRe = /Sent (Access-\w+)/;

export function summarize(text: string, format: Format): Record<string, string> {
  const out: Record<string, string> = {};
  const lines = text.split('\n');
  if (format === 'nps') {
    for (const line of lines) {
      const idx = line.indexOf(':');
      if (idx === -1) continue;
      const label = line.slice(0, idx).trim();
      const value = line.slice(idx + 1).trim();
      const key = npsFields[label];
      if (key && value !== '' && value !== '-' && !(key in out)) {
        out[key] = value;
      }
    }
  } else if (format === 'freeradius') {
    let errors = 0;
    let firstError = '';
    let finalReply = '';
    for (const line of lines) {
      if (errorLineRe.test(line)) {
        errors++;
        if (!firstError) firstError = line.trim();
      }
      const m = sentReplyRe.exec(line);
      if (m) finalReply = m[1];
    }
    if (errors > 0) {
      out['error_lines'] = String(errors);
      out['first_error'] = firstError;
    }
    if (finalReply) out['final_reply'] = finalReply;
  }
  return out;
}
