"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    Object.defineProperty(o, k2, { enumerable: true, get: function() { return m[k]; } });
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
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
const chalk_1 = __importDefault(require("chalk"));
const conf_1 = __importDefault(require("conf"));
const dotenv = __importStar(require("dotenv"));
const api_service_1 = require("./api-service");
dotenv.config({ path: __dirname + '/../.env' });
const config = new conf_1.default();
if (process.env.RESET_CONF == 'true') {
    console.log('Clearing config state, all messages will be exported');
    config.clear();
}
console.log(chalk_1.default.gray(`Using config path ${config.path}`));
let cntr = config.get('cntr') ? parseInt(config.get('cntr')) : 0;
cntr++;
console.log(chalk_1.default.greenBright(`Run cntr is ${cntr}`));
config.set('cntr', cntr);
const token = process.env.SLACK_API_TOKEN;
if (!token) {
    console.log(chalk_1.default.red('SLACK_API_TOKEN not set. Please set in environmental variables or .env file'));
    process.exit(0);
}
const targetFolder = process.env.EXPORT_ROOT;
if (!targetFolder) {
    console.log(chalk_1.default.red('EXPORT_ROOT not set. Please set in environmental variables or .env file'));
    process.exit(0);
}
(() => __awaiter(void 0, void 0, void 0, function* () {
    try {
        const service = new api_service_1.APIService(config, token, targetFolder, process.env.MIN_DATE_ISO, process.env.MAX_DATE_ISO, process.env.EXPORT_JSON === 'true', process.env.EXPORT_FILES === 'true');
        yield service.export();
    }
    catch (x) {
        console.log(chalk_1.default.red(`Error processing : ${x}`));
    }
}))();
