"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.APIService = void 0;
/* eslint-disable no-constant-condition */
/* eslint-disable @typescript-eslint/no-non-null-assertion */
const web_api_1 = require("@slack/web-api");
const chalk_1 = __importDefault(require("chalk"));
const date_fns_1 = require("date-fns");
const fs_1 = require("fs");
const path_1 = __importDefault(require("path"));
const sanitize_filename_1 = __importDefault(require("sanitize-filename"));
const axios_1 = __importDefault(require("axios"));
const timeout = (ms) => new Promise((res) => setTimeout(res, ms));
const REQUIRED_SCOPES = [
    'channels:history',
    'channels:read',
    'groups:history',
    'groups:read',
    'im:history',
    'im:read',
    'mpim:history',
    'mpim:read',
    'files:read',
    'users:read'
];
function delay(ms) {
    return __awaiter(this, void 0, void 0, function* () {
        yield timeout(ms);
    });
}
class APIService {
    constructor(state, _authToken, exportRoot, startDateISO, endDateISO, exportJson, exportFiles) {
        this.state = state;
        this.exportRoot = exportRoot;
        this.exportJson = exportJson;
        this.exportFiles = exportFiles;
        this.team = '';
        this.user = '';
        this.userMap = {};
        this.users = [];
        this.conversations = [];
        this.startDate = null;
        this.endDate = null;
        this.web = new web_api_1.WebClient(_authToken);
        if (startDateISO) {
            this.startDate = date_fns_1.parseISO(startDateISO);
        }
        if (endDateISO) {
            this.endDate = date_fns_1.parseISO(endDateISO);
        }
        this.axios = axios_1.default.create({
            headers: {
                'Authorization': 'Bearer ' + _authToken
            }
        });
    }
    export() {
        var _a;
        return __awaiter(this, void 0, void 0, function* () {
            try {
                yield fs_1.promises.access(this.exportRoot);
            }
            catch (x) {
                throw new Error(`EXPORT_TARGET_FOLDER value ${this.exportRoot} does not exist or is inaccessible: ${x}`);
            }
            const r = yield this.web.auth.test();
            if (!r.ok) {
                return;
            }
            this.team = r.team;
            this.user = r.user;
            console.log(`Operating on team "${this.team}" and user "${this.user}"`);
            console.log(`Writing output to ${this.exportRoot}`);
            const scopes = (_a = r.response_metadata) === null || _a === void 0 ? void 0 : _a.scopes;
            if (!scopes) {
                throw new Error('Empty scopes?');
            }
            for (const scope of REQUIRED_SCOPES) {
                if (scopes.indexOf(scope) < 0) {
                    throw new Error(`Missing required scope ${scope}`);
                }
            }
            console.log(chalk_1.default.greenBright('    Confirmed all necessary scopes present'));
            yield this.loadUsers();
            yield this.loadConversations();
            for (const conv of this.conversations) {
                yield this.writeNewMsgs(conv, this.startDate, this.endDate);
            }
        });
    }
    downloadFile(conv, file) {
        return __awaiter(this, void 0, void 0, function* () {
            const convTargetFolder = path_1.default.join(this.exportRoot, sanitize_filename_1.default(conv.name));
            try {
                yield fs_1.promises.access(convTargetFolder);
            }
            catch (x) {
                yield fs_1.promises.mkdir(convTargetFolder);
            }
            const target = path_1.default.join(convTargetFolder, `${file.id}-${file.name}`);
            const resp = yield this.axios.get(file.url_private, { responseType: 'arraybuffer' });
            yield fs_1.promises.writeFile(target, resp.data);
            console.log(chalk_1.default.blue(`           Downloaded ${target}`));
        });
    }
    loadUsers() {
        return __awaiter(this, void 0, void 0, function* () {
            const r = yield this.web.users.list();
            this.users = r.members;
            this.userMap = {};
            for (const u of this.users) {
                this.userMap[u.id] = u;
            }
        });
    }
    loadConversations() {
        return __awaiter(this, void 0, void 0, function* () {
            const r = yield this.web.conversations.list({ types: 'public_channel, private_channel, mpim, im' });
            this.conversations = r.channels;
            for (const c of this.conversations) {
                if (c.is_im) {
                    const user = this.users.find(x => c.user == x.id);
                    c.name = user === null || user === void 0 ? void 0 : user.name;
                    c.is_private = true;
                    c.type = 'd-msg';
                }
                else if (c.is_mpim) {
                    c.type = 'g-msg';
                }
                else {
                    if (c.is_private) {
                        c.type = 'pr-ch';
                    }
                    c.type = 'pu-ch';
                }
            }
        });
    }
    static dateToSlackTs(date) {
        const ms = date.getTime();
        const seconds = ms / 1000;
        return seconds.toFixed(6);
    }
    buildSimpleMsg(msg) {
        const ms = Math.floor(parseFloat(msg.ts) * 1000);
        const date = new Date(ms);
        //   const formatted = format(date, 'yyyy-MM-dd HH:mm:ss');    
        const name = this.userMap[msg.user] ? this.userMap[msg.user].name : 'unknown';
        return {
            name: name,
            ts: date,
            text: msg.text
        };
    }
    writeNewMsgs(conv, startDate, endDate) {
        return __awaiter(this, void 0, void 0, function* () {
            const lastIsoDate = this.state.get(conv.id);
            let oldestDate = null;
            if (lastIsoDate) {
                oldestDate = date_fns_1.parseISO(lastIsoDate);
                console.log(chalk_1.default.grey(`   ${conv.name} Using previously last saved date ${APIService.printDate(oldestDate)}`));
            }
            else if (endDate) {
                oldestDate = startDate;
            }
            const finishDate = yield this.writeMessages(conv, oldestDate, endDate);
            if (finishDate) {
                this.state.set(conv.id, date_fns_1.formatISO(finishDate));
            }
        });
    }
    static printDate(date) {
        if (!date) {
            return 'null';
        }
        return date_fns_1.format(date, 'yyyy-MM-dd HH:mm:ss');
    }
    writeMessages(conv, oldestDate, newestDate) {
        var _a;
        return __awaiter(this, void 0, void 0, function* () {
            let cursor = undefined;
            const msgs = [];
            const jsonMsgs = [];
            let lastDate = oldestDate;
            let extraPages = false;
            while (true) {
                if (extraPages) {
                    process.stdout.write('-');
                }
                const data = yield this.web.conversations.history({
                    channel: conv === null || conv === void 0 ? void 0 : conv.id,
                    latest: newestDate ? APIService.dateToSlackTs(newestDate) : undefined,
                    oldest: oldestDate ? APIService.dateToSlackTs(oldestDate) : undefined,
                    cursor: cursor,
                    limit: 1000
                });
                const slackMsgs = data.messages;
                for (const s of slackMsgs) {
                    const simpleMsg = this.buildSimpleMsg(s);
                    msgs.push(simpleMsg);
                    if (this.exportJson) {
                        jsonMsgs.push(s);
                    }
                    if (this.exportFiles) {
                        if (s.files && s.files.length > 0) {
                            for (const f of s.files) {
                                yield this.downloadFile(conv, f);
                            }
                        }
                    }
                }
                if (!data.has_more) {
                    break;
                }
                if (!extraPages) {
                    process.stdout.write('    ');
                    extraPages = true;
                }
                cursor = (_a = data.response_metadata) === null || _a === void 0 ? void 0 : _a.next_cursor;
            }
            if (extraPages) {
                process.stdout.write('\r\n');
            }
            msgs.sort((a, b) => {
                if (a.ts < b.ts) {
                    return -1;
                }
                if (a.ts > b.ts) {
                    return 1;
                }
                return 0;
            });
            if (msgs.length > 0) {
                yield this.writeMsgs(this.exportRoot, conv, msgs);
                lastDate = date_fns_1.add(msgs[msgs.length - 1].ts, { seconds: 1 });
                if (this.exportJson) {
                    const target = path_1.default.join(this.exportRoot, `${conv.type}-${sanitize_filename_1.default(conv.name)}_${date_fns_1.format(new Date(), 'yyyy-MM-dd-HHmmss')}.json`);
                    yield fs_1.promises.writeFile(target, JSON.stringify(jsonMsgs));
                }
            }
            else {
                // no messages found
            }
            console.log(chalk_1.default.yellow(`    Conv ${conv.name}: ${msgs.length}`));
            return lastDate;
        });
    }
    writeMsgs(root, conv, msgs) {
        return __awaiter(this, void 0, void 0, function* () {
            const lines = [];
            const target = path_1.default.join(root, `${conv.type}-${sanitize_filename_1.default(conv.name)}_${date_fns_1.format(new Date(), 'yyyy-MM-dd-HHmmss')}.txt`);
            for (const m of msgs) {
                const formatted = date_fns_1.format(m.ts, 'yyyy-MM-dd HH:mm:ss');
                const line = `${formatted} ${m.name}: ${m.text}`;
                lines.push(line);
            }
            yield fs_1.promises.writeFile(target, lines.join('\r\n'));
        });
    }
}
exports.APIService = APIService;
