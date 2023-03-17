import {ethers} from 'ethers';
import {SiweMessage} from 'siwe';
import {Buffer} from 'buffer';

// @ts-ignore
window.Buffer = Buffer;
const domain = window.location.host;
const origin = window.location.origin;

const provider = new ethers.providers.Web3Provider(
    (window as any).ethereum
);
const signer = provider.getSigner();

function createSiweMessage(address: string) {
    const message = new SiweMessage({
        domain: domain,
        address: address,
        uri: origin,
        version: '1',
        chainId: 1
    });
    return message.prepareMessage();
}

const addresses = await provider.listAccounts();
console.log(addresses)

async function displayLoginButton() {
    const address = await signer.getAddress()
    document.querySelector<HTMLDivElement>('#login')!.innerHTML = `<button id='siweBtn'>Sign-in with Ethereum with ${address}</button>`
    const signinBtn = document.getElementById('siweBtn');
    if (signinBtn != null) {
        signinBtn.onclick = signInWithEthereum;
    }
}

if (addresses.length == 0) {
    document.querySelector<HTMLDivElement>('#login')!.innerHTML = `<button id='connectWalletBtn'>Connect wallet</button>`
    const connectWalletBtn = document.getElementById('connectWalletBtn');
    if (connectWalletBtn != null) {
        connectWalletBtn.onclick = connectWallet;
    }
} else {
    await displayLoginButton();

}

async function connectWallet() {
    provider.send('eth_requestAccounts', [])
        .catch(() => console.log('user rejected request'));
    await displayLoginButton();
}

async function signInWithEthereum() {
    const message = createSiweMessage(
        await signer.getAddress()
    );

    const signature = await signer.signMessage(message)

    const signatureField = document.getElementById('signature') as HTMLInputElement
    if (signatureField != null) {
        signatureField.value = signature
        console.log(signatureField)
    }
    const messageField = document.getElementById('message') as HTMLInputElement
    if (messageField != null) {
        var result = "";
        for (var i = 0; i < message.length; i++) {
            var hex = message.charCodeAt(i).toString(16);
            result += ("000" + hex).slice(-2);
        }
        messageField.value = result
    }
    const form = document.getElementById('loginform') as HTMLFormElement
    if (form != null) {
        console.log(form)
        form.submit()
    }
}
