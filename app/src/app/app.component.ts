import { Component } from '@angular/core';
import { Platform } from '@ionic/angular';
import { SplashScreen } from '@ionic-native/splash-screen/ngx';
import { StatusBar } from '@ionic-native/status-bar/ngx';
//import { FCM } from '@ionic-native/fcm/ngx';
import { FCM } from "cordova-plugin-fcm-with-dependecy-updated/ionic";


@Component({
  selector: 'app-root',
  templateUrl: 'app.component.html',
  styleUrls: ['app.component.scss']
})

export class AppComponent {
  constructor(
      private platform: Platform,
      private splashScreen: SplashScreen,
      private statusBar: StatusBar,
  ) {
    this.initializeApp();
  }
  initializeApp() {
    this.platform.ready().then(() => {
      this.statusBar.styleDefault();
      this.splashScreen.hide();
      // subscribe to a topic
      FCM.subscribeToTopic('earthquake');
      // get FCM token
      FCM.getToken().then(token => {
        console.log(token);
      });
      // ionic push notification example
      FCM.onNotification().subscribe(data => {
        console.log(data);
        if (data.wasTapped) {
          console.log('Received in background');
        } else {
          console.log('Received in foreground');
        }
      });
      // refresh the FCM token
      FCM.onTokenRefresh().subscribe(token => {
        console.log(token);
      });
      // unsubscribe from a topic
      // this.fcm.unsubscribeFromTopic('offers');
    });

  }
}