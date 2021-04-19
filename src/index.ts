import chalk from 'chalk';
import Conf from 'conf';
import * as dotenv from 'dotenv';
import { APIService } from './api-service';

dotenv.config({ path: __dirname+'/../.env' });
const config = new Conf();

if (process.env.RESET_CONF=='true') {
    console.log('Clearing config state, all messages will be exported');
    config.clear();
}
console.log(chalk.gray(`Using config path ${config.path}`));
let cntr = config.get('cntr')? parseInt(config.get('cntr') as string): 0 ;
cntr++;
console.log(chalk.greenBright(`Run cntr is ${cntr}`));
config.set('cntr', cntr);


const token = process.env.SLACK_API_TOKEN;
if (!token) {
    console.log(chalk.red('SLACK_API_TOKEN not set. Please set in environmental variables or .env file'));
    process.exit(0);
}
const targetFolder = process.env.EXPORT_ROOT;
if (!targetFolder) {
    console.log(chalk.red('EXPORT_ROOT not set. Please set in environmental variables or .env file'));
    process.exit(0);
}
(async () => {

    try {
        const service = new APIService(config, token, targetFolder, process.env.MIN_DATE_ISO, 
            process.env.MAX_DATE_ISO, 
            process.env.EXPORT_JSON==='true',
            process.env.EXPORT_FILES==='true');
        await service.export();
    } catch (x) {
        console.log(chalk.red(`Error processing : ${x}`));
    }

})();