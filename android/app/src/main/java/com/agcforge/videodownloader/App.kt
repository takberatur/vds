package com.agcforge.videodownloader

import android.annotation.SuppressLint
import android.app.Application
import com.agcforge.videodownloader.utils.DownloadManagerCleaner

class App : Application() {

    companion object {
        private const val TAG = "App_Root_Main"

        @SuppressLint("StaticFieldLeak")
        @Volatile
        private var mInstance: App? = null

        fun getInstance(): App {
            return mInstance ?: throw IllegalStateException("App not initialized")
        }

        var click = 3
        var backclick = 3

        var AdsClickCount = 0
        var backAdsClickCount = 0

        var TypeAds = "admob"

        const val Onesignal_ID = "86149c37-f641-4b38-a75a-5f52df523c07"


    }
    override fun onCreate() {
        super.onCreate()
        mInstance = this
		DownloadManagerCleaner.clearFailedDownloads(this)
    }
}
